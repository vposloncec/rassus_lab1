# frozen_string_literal: true

class Sensor < ApplicationRecord
  has_many :readings

  scope :all_except, ->(user_id) { where.not(id: user_id) }

  def nearest_neighbour
    neighbours.first&.keys&.first
  end

  def neighbours
    distance_to_each.sort_by(&:values)
  end

  def location
    @location ||= OpenStruct.new(lat: latitude, lon: longitude)
  end

  def distance_to_each
    Sensor.all_except(id).map do |other_sensor|
      other_sensor_id = other_sensor.id
      { other_sensor_id => calculate_distance(location, other_sensor.location) }
    end
  end

  def calculate_distance(loc1, loc2)
    # Haversine formula ?? (magic)
    r = 6371
    dlon = loc2.lon - loc1.lon
    dlat = loc2.lat - loc1.lat
    a = Math.sin(dlat / 2)**2 + Math.cos(loc1.lat) * Math.cos(loc2.lat) * Math.sin(dlon / 2)**2
    c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
    r * c
  end
end
