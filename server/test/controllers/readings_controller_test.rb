# frozen_string_literal: true

require 'test_helper'

class ReadingsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @reading = readings(:one)
  end

  test 'should get index' do
    get readings_url, as: :json
    assert_response :success
  end

  test 'should create reading' do
    assert_difference('Reading.count') do
      post readings_url,
           params: { reading: { co: @reading.co, float: @reading.float, humidity: @reading.humidity, pressure: @reading.pressure, so2: @reading.so2, temperature: @reading.temperature } }, as: :json
    end

    assert_response 201
  end

  test 'should show reading' do
    get reading_url(@reading), as: :json
    assert_response :success
  end

  test 'should update reading' do
    patch reading_url(@reading),
          params: { reading: { co: @reading.co, float: @reading.float, humidity: @reading.humidity, pressure: @reading.pressure, so2: @reading.so2, temperature: @reading.temperature } }, as: :json
    assert_response 200
  end

  test 'should destroy reading' do
    assert_difference('Reading.count', -1) do
      delete reading_url(@reading), as: :json
    end

    assert_response 204
  end
end
