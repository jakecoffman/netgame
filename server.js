var Server = function(canvas, status) {
  // Connected clients and their entities.
  this.clients = [];
  this.entities = [];

  // Last processed input for each client.
  this.last_processed_input = [];

  // Simulated network connection.
  this.network = new LagNetwork();

  // UI.
  this.canvas = canvas;
  this.context = canvas.getContext('2d')
  this.status = status;

  // Default update rate.
  this.setUpdateRate(10);
}

Server.prototype.connect = function(client) {
  // Give the Client enough data to identify itself.
  client.server = this;
  client.entity_id = this.clients.length;
  this.clients.push(client);

  // Create a new Entity for this Client.
  var entity = new Entity();
  this.entities.push(entity);
  entity.entity_id = client.entity_id;

  // Set the initial state of the Entity (e.g. spawn point)
  var spawn_points = [4, 6];
  entity.x = spawn_points[client.entity_id];
}

Server.prototype.setUpdateRate = function(hz) {
  this.update_rate = hz;

  clearInterval(this.update_interval);
  this.update_interval = setInterval(()=>this.update(), 1000 / this.update_rate);
}

Server.prototype.update = function() {
  this.processInputs();
  this.sendWorldState();
  renderWorld(this.context, this.canvas, this.entities);
}


// Check whether this input seems to be valid (e.g. "make sense" according
// to the physical rules of the World)
Server.prototype.validateInput = function(input) {
  return Math.abs(input.press_time) <= 1 / 40;
}


Server.prototype.processInputs = function() {
  // Process all pending messages from clients.
  while (true) {
    var message = this.network.receive();
    if (!message) {
      break;
    }

    // Update the state of the entity, based on its input.
    // We just ignore inputs that don't look valid; this is what prevents clients from cheating.
    if (this.validateInput(message)) {
      var id = message.entity_id;
      this.entities[id].applyInput(message);
      this.last_processed_input[id] = message.input_sequence_number;
    }

  }

  // Show some info.
  var info = "Last acknowledged input: ";
  for (var i = 0; i < this.clients.length; ++i) {
    info += "Player " + i + ": #" + (this.last_processed_input[i] || 0) + "   ";
  }
  this.status.textContent = info;
}


// Send the world state to all the connected clients.
Server.prototype.sendWorldState = function() {
  // Gather the state of the world. In a real app, state could be filtered to avoid leaking data
  // (e.g. position of invisible enemies).
  var world_state = [];
  var num_clients = this.clients.length;
  for (let i = 0; i < num_clients; i++) {
    var entity = this.entities[i];
    world_state.push({entity_id: entity.entity_id,
      position: entity.x,
      last_processed_input: this.last_processed_input[i]});
  }

  // Broadcast the state to all the clients.
  for (let i = 0; i < num_clients; i++) {
    var client = this.clients[i];
    client.network.send(client.lag, world_state);
  }
}
