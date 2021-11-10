# frozen_string_literal: true

class NearestController < ApplicationController
  before_action :set_sensor, only: [:show]

  def show
    render json: Sensor.find(@sensor.nearest_neighbour)
  end

  private

  def set_sensor
    @sensor = Sensor.find(params[:sensor_id])
  end
end
