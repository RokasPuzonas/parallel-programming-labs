require "json"
$VERBOSE = nil  # Suppress verbose warnings

# Define a class for data entries
class DataEntry
  attr_reader :name
  attr_reader :sugar
  attr_reader :criteria

  def initialize(name, sugar, criteria)
    @name = name
    @sugar = sugar
    @criteria = criteria
  end

  def to_s
    "#<DataEntry:#{@name}, #{@sugar}, #{@criteria}>"
  end
end

# Define a class for result entries
class ResultEntry
  attr_reader :name
  attr_reader :sugar

  def initialize(name, sugar)
    @name = name
    @sugar = sugar
  end

  def to_s
    "#<ResultEntry:#{@name}, #{@sugar}>"
  end
end

# Define a class for input messages
class InputMessage
  attr_reader :data

  def initialize(data)
    @data = data
  end
end

# Define a class for result messages
class ResultMessage
  attr_reader :tid
  attr_reader :data

  def initialize(tid, data)
    @tid = tid
    @data = data
  end
end

# Define a class for finished messages
class FinishedMessage
  attr_reader :tid

  def initialize(tid)
    @tid = tid
  end
end

# Define a class for final results messages
class FinalResultsMessage
  attr_reader :data

  def initialize(data)
    @data = data
  end
end

# Method to read data from a file and convert it to DataEntry objects
def read_data(file_path)
  data = []
  json_data = File.read(file_path)
  JSON.parse(json_data).each do |entry|
    data.push(DataEntry.new(entry["name"], entry["sugar"].to_f, entry["criteria"].to_i))
  end
  return data
end

# Constants
worker_count = 12
data = read_data("IF-1-1_PuzonasR_dat_1.json")
results_path = "results.txt"
logs_path = "logs.txt"

# Method to process data and return a ResultMessage or FinishedMessage
def process_data(tid, data_entry)
  if data_entry.sugar > data_entry.criteria
    sleep 0.25
    ResultMessage.new(tid, ResultEntry.new(data_entry.name, data_entry.sugar))
  else
    FinishedMessage.new(tid)
  end
end

# Create Ractors for workers
workers = Array.new(worker_count) do |index|
  Ractor.new(index) do |index|
    distributer = Ractor.receive

    loop do
      data_entry = Ractor.receive
      if data_entry == nil then break end

      distributer.send(process_data(index, data_entry))
    end
  end
end

# Create Ractors for result collector, results writer, and logger
result_collector = Ractor.new do
  results = []
  distributer = Ractor.receive

  loop do
    result = Ractor.receive
    if result == nil then break end

    index = results.index { |old_value| old_value.sugar > result.sugar }
    if index then
      results.insert(index, result)
    else
      results.push(result)
    end
  end

  distributer.send(FinalResultsMessage.new(results))
end

results_writer = Ractor.new(results_path) do |results_path|
  results = Ractor.receive
  File.open(results_path, "w") do |file|
    file.puts "--------------------------------"
    file.puts "| %-15s | %-10s |" % ["Name", "Sugar"]
    file.puts "--------------------------------"
    results.each do |result|
      file.puts "| %-15s | %-10s |" % [result.name, result.sugar]
    end
    file.puts "--------------------------------"
  end
end

logger = Ractor.new(logs_path) do |logs_path|
  File.open(logs_path, "w") do |file|
    loop do
      msg = Ractor.receive
      if msg == nil then break end
      file.puts msg
    end
  end
end

# Create a distributer Ractor to manage communication between other Ractors
distributer = Ractor.new(logger, result_collector, results_writer, workers, worker_count) do |logger, result_collector, results_writer, workers, worker_count|
  workers.each do |worker|
    worker.send(Ractor.current)
  end
  result_collector.send(Ractor.current)

  worker_counter = 0
  input_counter = 0
  result_counter = 0

  workers_active = true

  loop do
    msg = Ractor.receive
    if msg == nil then break end

    if workers_active and msg.is_a?(InputMessage) then
      tid = input_counter % worker_count
      logger.send("[                ->worker#{tid.to_s.ljust(2)}        ] Input message, data: #{msg.data}")
      workers[tid].send(msg.data)
      input_counter += 1
    elsif workers_active and msg.is_a?(ResultMessage) then
      logger.send("[worker#{msg.tid.to_s.ljust(2)}        ->result_collector] Result message, data: #{msg.data}")
      result_counter += 1
      result_collector.send(msg.data)
    elsif workers_active and msg.is_a?(FinishedMessage) then
      logger.send("[worker#{msg.tid.to_s.ljust(2)}        ->                ] Finished message")
      result_counter += 1
    elsif msg.is_a?(FinalResultsMessage) then
      logger.send("[result_collector->result_writer   ] Writing results to file")
      results_writer.send(msg.data)
    else
      logger.send("[                ->                ] Unknown message")
    end

    if input_counter == result_counter then
      logger.send("[                ->result_collector] Stop message")
      result_collector.send(nil)
      workers.each_with_index do |worker, tid|
        logger.send("[                ->worker#{tid.to_s.ljust(2)}        ] Stop message")
        worker.send(nil)
      end
    end
  end

  logger.send("[                ->logger          ] Stop message")
  logger.send(nil)
end

# Send input messages to the distributer
data.each do |data_entry|
  distributer.send(InputMessage.new(data_entry))
end

# Finalize Ractors by sending stop messages and collecting results
distributer.take
workers.each(&:take)
result_collector.take
results_writer.take
logger.take

# when 'worker_count = 1' , time taken: 10.058
# when 'worker_count = 12', time taken: 1.055
