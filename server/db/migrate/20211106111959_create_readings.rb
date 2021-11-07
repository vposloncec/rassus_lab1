class CreateReadings < ActiveRecord::Migration[6.1]
  def change
    create_table :readings do |t|
      t.float :temperature
      t.float :pressure
      t.float :humidity
      t.float :co
      t.string :so2

      t.belongs_to :sensor

      t.timestamps
    end
  end
end
