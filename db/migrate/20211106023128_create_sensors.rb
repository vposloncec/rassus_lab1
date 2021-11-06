class CreateSensors < ActiveRecord::Migration[6.1]
  def change
    create_table :sensors do |t|
      t.float :latitude
      t.float :longitude
      t.string :ip
      t.integer :port

      t.timestamps
    end
  end
end
