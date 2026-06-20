#!/usr/bin/env ruby
# frozen_string_literal: true

require "json"
require "pathname"
require "yaml"

ROOT = Pathname.new(__dir__).join("..", "openapi", "openapi.yaml").cleanpath
HTTP_METHODS = %w[get post put patch delete options head trace].freeze
REQUIRED_OPERATIONS = {
  "GET /api/v1/health" => "getHealth",
  "GET /api/v1/ready" => "getReadiness",
  "POST /api/v1/auth/register" => "register",
  "POST /api/v1/auth/login" => "login",
  "POST /api/v1/auth/refresh" => "refreshSession",
  "POST /api/v1/auth/logout" => "logout",
  "GET /api/v1/assets/search" => "searchAssets",
  "GET /api/v1/assets/{ticker}" => "getAsset",
  "GET /api/v1/portfolios" => "listPortfolios",
  "POST /api/v1/portfolios" => "createPortfolio",
  "GET /api/v1/portfolios/{portfolioId}" => "getPortfolio",
  "PATCH /api/v1/portfolios/{portfolioId}" => "updatePortfolio",
  "DELETE /api/v1/portfolios/{portfolioId}" => "deletePortfolio",
  "GET /api/v1/portfolios/{portfolioId}/summary" => "getPortfolioSummary",
  "GET /api/v1/portfolios/{portfolioId}/snapshots" => "listPortfolioSnapshots",
  "GET /api/v1/portfolios/{portfolioId}/transactions" => "listTransactions",
  "POST /api/v1/portfolios/{portfolioId}/transactions" => "createTransaction",
  "PATCH /api/v1/portfolios/{portfolioId}/transactions/{transactionId}" => "correctTransaction",
  "DELETE /api/v1/portfolios/{portfolioId}/transactions/{transactionId}" => "reverseTransaction",
  "GET /api/v1/dividends/calendar" => "getDividendCalendar",
  "POST /api/v1/dividends/calculate" => "calculateDividend",
  "GET /api/v1/dashboard" => "getDashboard"
}.freeze
PUBLIC_OPERATIONS = %w[getHealth getReadiness register login searchAssets getAsset calculateDividend].freeze
REFRESH_OPERATIONS = %w[refreshSession logout].freeze
IDEMPOTENT_OPERATIONS = %w[
  createPortfolio deletePortfolio createTransaction correctTransaction reverseTransaction calculateDividend
].freeze
REQUIRED_SCHEMAS = %w[
  Money Decimal BusinessDate SystemTimestamp Asset AssetType Portfolio Transaction
  TransactionType PortfolioSummary PortfolioSnapshot DividendEvent DividendCalculation
  RealReturn PurchasingPower Pagination Error BaseResponse ErrorResponse
].freeze

class ContractValidator
  attr_reader :reference_count, :operation_count

  def initialize(root)
    @root = root.expand_path
    @documents = {}
    @visited_references = {}
    @reference_count = 0
    @operation_count = 0
    @errors = []
  end

  def validate!
    root_document = load_document(@root)
    validate_structure(root_document)
    validate_financial_guard_vectors
    walk(root_document, @root, "#")

    unless @errors.empty?
      warn @errors.map { |error| "ERROR: #{error}" }.join("\n")
      exit 1
    end

    puts "OpenAPI validation passed: #{@operation_count} operations, #{@reference_count} resolved references, #{@documents.length} documents"
  end

  private

  def validate_financial_guard_vectors
    schemas_path = @root.dirname.join("components", "schemas.yaml")
    schemas = load_document(schemas_path)
    transaction = schemas.fetch("Transaction", {}).fetch("properties", {})

    reject_vector("Transaction.quantity negative", "-1.00000000", transaction["quantity"], schemas_path)
    reject_vector("Transaction.grossAmount negative", { "amount" => "-1.00000000", "currency" => "RUB" },
                  transaction["grossAmount"], schemas_path)
    transaction_schema = schemas.fetch("Transaction")
    transaction_example = load_document(@root.dirname.join("examples", "transactions.json"))
                          .dig("createResponse", "value", "data")
    invalid_buy = Marshal.load(Marshal.dump(transaction_example)).merge("quantity" => nil)
    reject_vector("Transaction BUY without quantity", invalid_buy, transaction_schema, schemas_path)
    invalid_cash = Marshal.load(Marshal.dump(transaction_example)).merge("transactionType" => "DEPOSIT")
    reject_vector("Transaction DEPOSIT with asset fields", invalid_cash, transaction_schema, schemas_path)

    calculator = schemas.fetch("DividendCalculationRequest", {}).fetch("properties", {})
    reject_vector("DividendCalculationRequest.quantity zero", "0.00000000", calculator["quantity"], schemas_path)
    reject_vector("DividendCalculationRequest.quantity negative", "-1.00000000", calculator["quantity"], schemas_path)
    reject_vector("DividendCalculationRequest.dividendPerUnit negative",
                  { "amount" => "-1.00000000", "currency" => "RUB" }, calculator["dividendPerUnit"], schemas_path)
    reject_vector("DividendCalculationRequest.positionCost zero",
                  { "amount" => "0.00000000", "currency" => "RUB" }, calculator["positionCost"], schemas_path)

    result_schema = schemas.fetch("DividendCalculation")
    result_example = load_document(@root.dirname.join("examples", "dividends.json"))
                     .dig("calculateResponse", "value", "data")
    yield_without_cost = Marshal.load(Marshal.dump(result_example)).merge("positionCost" => nil)
    reject_vector("DividendCalculation yield without position cost", yield_without_cost, result_schema, schemas_path)
    missing_yield_with_cost = Marshal.load(Marshal.dump(result_example)).merge("grossYield" => nil)
    reject_vector("DividendCalculation null yield with position cost", missing_yield_with_cost, result_schema, schemas_path)
  end

  def reject_vector(name, value, schema, schema_path)
    @errors << "invalid financial vector accepted: #{name}" if schema_matches?(value, schema, schema_path)
  end

  def load_document(path)
    normalized = path.expand_path
    return @documents.fetch(normalized) if @documents.key?(normalized)

    content = normalized.read
    document = if normalized.extname == ".json"
                 JSON.parse(content)
               else
                 YAML.safe_load(content, aliases: true)
               end
    @documents[normalized] = document
  rescue Errno::ENOENT
    @errors << "missing referenced file #{normalized}"
    {}
  rescue JSON::ParserError, Psych::SyntaxError => error
    @errors << "cannot parse #{normalized}: #{error.message}"
    {}
  end

  def validate_structure(document)
    @errors << "root openapi version must be 3.1.0" unless document["openapi"] == "3.1.0"
    @root_security = document["security"]
    @errors << "root security must require bearerAuth" unless @root_security == [{ "bearerAuth" => [] }]
    paths = document.fetch("paths", {})

    missing_paths = REQUIRED_OPERATIONS.keys.map { |key| key.split(" ", 2).last }.uniq - paths.keys
    extra_unversioned = paths.keys.reject { |path| path.start_with?("/api/v1/") }
    @errors << "missing required paths: #{missing_paths.join(', ')}" unless missing_paths.empty?
    @errors << "unversioned paths: #{extra_unversioned.join(', ')}" unless extra_unversioned.empty?

    REQUIRED_SCHEMAS.each do |schema_name|
      unless document.dig("components", "schemas", schema_name)
        @errors << "missing registered canonical schema #{schema_name}"
      end
    end

    operation_ids = []
    actual_operations = {}
    paths.each do |path, path_item|
      next unless path_item.is_a?(Hash)

      HTTP_METHODS.each do |method|
        operation = path_item[method]
        next unless operation

        @operation_count += 1
        operation_id = operation["operationId"]
        @errors << "#{method.upcase} #{path} has no operationId" unless operation_id
        operation_ids << operation_id if operation_id
        actual_operations["#{method.upcase} #{path}"] = operation_id
        validate_operation(path, method, operation, Array(path_item["parameters"]))
      end
    end

    duplicates = operation_ids.group_by(&:itself).select { |_id, values| values.length > 1 }.keys
    @errors << "duplicate operationIds: #{duplicates.join(', ')}" unless duplicates.empty?
    missing = REQUIRED_OPERATIONS.keys - actual_operations.keys
    extras = actual_operations.keys - REQUIRED_OPERATIONS.keys
    @errors << "missing required operations: #{missing.join(', ')}" unless missing.empty?
    @errors << "unexpected operations: #{extras.join(', ')}" unless extras.empty?
    REQUIRED_OPERATIONS.each do |key, operation_id|
      actual = actual_operations[key]
      @errors << "#{key} must use operationId #{operation_id}, got #{actual}" if actual && actual != operation_id
    end
  end

  def validate_operation(path, method, operation, path_parameters)
    operation_id = operation["operationId"]
    validate_security(path, method, operation, operation_id)
    validate_idempotency(path, method, operation, operation_id, path_parameters)
    responses = operation.fetch("responses", {})
    success = responses.find { |status, _response| status.match?(/^2[0-9][0-9]$/) }
    unless success
      @errors << "#{method.upcase} #{path} has no 2xx response"
      return
    end

    resolved_response, response_path = dereference_with_path(success.last, @root)
    examples = resolved_response.dig("content", "application/json", "examples")
    @errors << "#{method.upcase} #{path} has no success example" unless examples.is_a?(Hash) && !examples.empty?
    success_schema = resolved_response.dig("content", "application/json", "schema")
    validate_envelope(success_schema, response_path, "BaseResponse", "#{method.upcase} #{path} success")
    validate_examples(examples, success_schema, response_path, "#{method.upcase} #{path} success")

    responses.each do |status, response|
      next if status.match?(/^2[0-9][0-9]$/)

      resolved_error, error_path = dereference_with_path(response, @root)
      error_schema = resolved_error.dig("content", "application/json", "schema")
      next unless error_schema

      error_location = "#{method.upcase} #{path} #{status}"
      validate_envelope(error_schema, error_path, "ErrorResponse", error_location)
      validate_examples(resolved_error.dig("content", "application/json", "examples"), error_schema, error_path, error_location)
    end

    request_body = operation["requestBody"]
    return unless request_body

    resolved_body, body_path = dereference_with_path(request_body, @root)
    body_examples = resolved_body.dig("content", "application/json", "examples")
    unless body_examples.is_a?(Hash) && !body_examples.empty?
      @errors << "#{method.upcase} #{path} request body has no example"
    end
    body_schema = resolved_body.dig("content", "application/json", "schema")
    validate_examples(body_examples, body_schema, body_path, "#{method.upcase} #{path} request")
  end

  def validate_security(path, method, operation, operation_id)
    security = operation.key?("security") ? operation["security"] : @root_security
    expected = if PUBLIC_OPERATIONS.include?(operation_id)
                 []
               elsif REFRESH_OPERATIONS.include?(operation_id)
                 [{ "refreshCookie" => [] }]
               else
                 [{ "bearerAuth" => [] }]
               end
    @errors << "#{method.upcase} #{path} has incorrect security declaration" unless security == expected
  end

  def validate_idempotency(path, method, operation, operation_id, path_parameters)
    parameters = path_parameters + Array(operation["parameters"])
    idempotency_parameters = parameters.filter do |parameter|
      parameter.is_a?(Hash) && parameter["$ref"]&.end_with?("/IdempotencyKey")
    end
    has_key = !idempotency_parameters.empty?
    idempotency_parameters.each do |parameter|
      resolved, = dereference_with_path(parameter, @root)
      @errors << "#{method.upcase} #{path} Idempotency-Key must be required" unless resolved["required"] == true
    end
    required = IDEMPOTENT_OPERATIONS.include?(operation_id)
    @errors << "#{method.upcase} #{path} must require Idempotency-Key" if required && !has_key
    @errors << "#{method.upcase} #{path} unexpectedly requires Idempotency-Key" if !required && has_key
  end

  def validate_envelope(schema, schema_path, expected, location)
    return @errors << "#{location} has no JSON schema" unless schema

    names = referenced_schema_names(schema, schema_path)
    @errors << "#{location} does not inherit #{expected}" unless names.include?(expected)
  end

  def referenced_schema_names(node, current_path, names = [])
    case node
    when Hash
      if node["$ref"]
        names << node["$ref"].split("/").last
        target, target_path, = resolve_reference(current_path, node["$ref"])
        referenced_schema_names(target, target_path, names)
      else
        node.each_value { |value| referenced_schema_names(value, current_path, names) }
      end
    when Array
      node.each { |value| referenced_schema_names(value, current_path, names) }
    end
    names
  end

  def validate_examples(examples, schema, schema_path, location)
    return unless examples.is_a?(Hash) && schema

    examples.each do |name, example|
      resolved_example, = dereference_with_path(example, schema_path)
      value = resolved_example.is_a?(Hash) && resolved_example.key?("value") ? resolved_example["value"] : resolved_example
      validate_instance(value, schema, schema_path, "#{location} example #{name}")
    end
  end

  def validate_instance(value, schema, schema_path, location)
    resolved, resolved_path = dereference_with_path(schema, schema_path)
    Array(resolved["allOf"]).each { |part| validate_instance(value, part, resolved_path, location) }
    if resolved["if"]
      branch = schema_matches?(value, resolved["if"], resolved_path) ? resolved["then"] : resolved["else"]
      validate_instance(value, branch, resolved_path, location) if branch
    end
    if resolved["oneOf"]
      matches = resolved["oneOf"].count { |part| schema_matches?(value, part, resolved_path) }
      @errors << "#{location} must match exactly one oneOf branch (matched #{matches})" unless matches == 1
      return
    end

    types = Array(resolved["type"])
    if !types.empty? && !types.any? { |type| type_matches?(value, type) }
      @errors << "#{location} has type #{value.class}, expected #{types.join(' or ')}"
      return
    end
    @errors << "#{location} is not in enum" if resolved["enum"] && !resolved["enum"].include?(value)
    @errors << "#{location} does not equal const" if resolved.key?("const") && resolved["const"] != value

    if value.is_a?(Hash)
      missing = Array(resolved["required"]) - value.keys
      @errors << "#{location} misses required properties: #{missing.join(', ')}" unless missing.empty?
      properties = resolved.fetch("properties", {})
      if resolved["additionalProperties"] == false
        extras = value.keys - properties.keys
        @errors << "#{location} has unknown properties: #{extras.join(', ')}" unless extras.empty?
      end
      value.each { |key, child| validate_instance(child, properties[key], resolved_path, "#{location}.#{key}") if properties[key] }
      if resolved["minProperties"] && value.length < resolved["minProperties"]
        @errors << "#{location} has fewer than #{resolved['minProperties']} properties"
      end
    elsif value.is_a?(Array)
      value.each_with_index { |child, index| validate_instance(child, resolved["items"], resolved_path, "#{location}[#{index}]") } if resolved["items"]
    elsif value.is_a?(String)
      @errors << "#{location} is shorter than minLength" if resolved["minLength"] && value.length < resolved["minLength"]
      @errors << "#{location} is longer than maxLength" if resolved["maxLength"] && value.length > resolved["maxLength"]
      @errors << "#{location} does not match pattern" if resolved["pattern"] && !Regexp.new(resolved["pattern"]).match?(value)
    elsif value.is_a?(Numeric)
      @errors << "#{location} is below minimum" if resolved["minimum"] && value < resolved["minimum"]
      @errors << "#{location} is above maximum" if resolved["maximum"] && value > resolved["maximum"]
    end
    if resolved["not"] && schema_matches?(value, resolved["not"], resolved_path)
      @errors << "#{location} matches a forbidden schema"
    end
  end

  def schema_matches?(value, schema, schema_path)
    before = @errors.length
    validate_instance(value, schema, schema_path, "candidate")
    matched = @errors.length == before
    @errors.slice!(before..-1) unless matched
    matched
  end

  def type_matches?(value, type)
    { "object" => Hash, "array" => Array, "string" => String, "integer" => Integer,
      "number" => Numeric, "boolean" => [TrueClass, FalseClass], "null" => NilClass }.fetch(type).then do |classes|
      Array(classes).any? { |klass| value.is_a?(klass) }
    end
  end

  def dereference(object, current_path)
    return object unless object.is_a?(Hash) && object["$ref"]

    target, target_path, = resolve_reference(current_path, object["$ref"])
    dereference(target, target_path)
  end

  def dereference_with_path(object, current_path)
    return [object, current_path] unless object.is_a?(Hash) && object["$ref"]

    target, target_path, = resolve_reference(current_path, object["$ref"])
    dereference_with_path(target, target_path)
  end

  def walk(node, current_path, location)
    case node
    when Hash
      if node["$ref"]
        target, target_path, key = resolve_reference(current_path, node["$ref"])
        unless @visited_references[key]
          @visited_references[key] = true
          walk(target, target_path, key)
        end
      end
      node.each do |key, value|
        next if key == "$ref"

        walk(value, current_path, "#{location}/#{key}")
      end
    when Array
      node.each_with_index { |value, index| walk(value, current_path, "#{location}/#{index}") }
    end
  end

  def resolve_reference(current_path, reference)
    @reference_count += 1
    file_part, fragment = reference.split("#", 2)
    target_path = file_part.nil? || file_part.empty? ? current_path : current_path.dirname.join(file_part).cleanpath
    document = load_document(target_path)
    pointer = fragment.nil? || fragment.empty? ? [] : fragment.sub(%r{^/}, "").split("/")
    target = pointer.reduce(document) do |value, raw_token|
      token = raw_token.gsub("~1", "/").gsub("~0", "~")
      if value.is_a?(Hash) && value.key?(token)
        value[token]
      else
        @errors << "unresolved reference #{reference} from #{current_path}"
        break({})
      end
    end
    key = "#{target_path.expand_path}##{fragment}"
    [target, target_path.expand_path, key]
  end
end

ContractValidator.new(ROOT).validate!
