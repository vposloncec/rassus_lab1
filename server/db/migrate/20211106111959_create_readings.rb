# frozen_string_literal: true

class CreateReadings < ActiveRecord::Migration[6.1]
  def change
    create_table :readings do |t|
      t.float :temperature
      t.float :pressure
      t.float :humidity
      t.float :co
      t.float :so2
      t.float :no2

      t.belongs_to :sensor

      t.timestamps
    end
  end
end
