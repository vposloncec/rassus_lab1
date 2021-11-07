class ReadingsController < ApplicationController
  before_action :set_reading, only: [:show, :update, :destroy]

  # GET /readings
  def index
    if params[:sensor_id]
      @readings = Sensor.find(params[:sensor_id]).readings
    else
      @readings = Reading.all
    end

    if @readings.empty?
      render status: :no_content
    else
      render json: @readings, except: [:created_at, :updated_at]
    end
  end

  # GET /readings/1
  def show
    render json: @reading
  end

  # POST /readings
  def create
    @reading = Reading.new(reading_params.merge({'sensor_id': params[:sensor_id]}))

    if @reading.save
      render json: @reading, status: :created, location: sensor_readings_url(@reading)
    else
      render json: @reading.errors, status: :unprocessable_entity
    end
  end

  # PATCH/PUT /readings/1
  def update
    if @reading.update(reading_params)
      render json: @reading
    else
      render json: @reading.errors, status: :unprocessable_entity
    end
  end

  # DELETE /readings/1
  def destroy
    @reading.destroy
  end

  private
  # Use callbacks to share common setup or constraints between actions.
  def set_reading
    @reading = Reading.find(params[:id])
  end

  # Only allow a list of trusted parameters through.
  def reading_params
    params.require(:reading).permit(:temperature, :pressure, :humidity, :co, :so2)
  end

end
