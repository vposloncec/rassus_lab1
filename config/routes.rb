Rails.application.routes.draw do
  resources :readings
  resources :sensors
  post '/register', to: 'sensors#create'
  get '/nearest/:sensor_id', to: 'nearest#show'
  # For details on the DSL available within this file, see https://guides.rubyonrails.org/routing.html
end
