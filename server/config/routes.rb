Rails.application.routes.draw do

  resources :sensors do
    resources :readings
    get '/nearest', to: 'nearest#show'
  end
  post '/register', to: 'sensors#create'
  # For details on the DSL available within this file, see https://guides.rubyonrails.org/routing.html
  # Not needed for the task itself, but user might find it useful
  resources :readings
end
