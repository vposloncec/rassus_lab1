require 'sinatra'
require 'pry'

before do
  @req_raw_body = request.body.read
  begin
  @data = JSON.parse(@req_raw_body)
  rescue { status 400 }
  request.body.rewind
  content_type :json
end

get '/' do
  'Welcome'
end

post '/register' do
  logger.info "Received request, #{@data}"
end

