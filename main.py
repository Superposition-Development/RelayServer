from flask import request
from init import app, cors
import database
from routes.auth import bpAuth
from routes.server import bpServer
from routes.channel import bpChannel

@app.before_request
def before_request(): 
    allowed_origin = request.headers.get('Origin')
    # if allowed_origin in ['http://localhost:4100']:
    cors._origins = allowed_origin

@app.route("/")
def home():
    return "eventually lets put something cool here"

app.register_blueprint(bpAuth)
app.register_blueprint(bpServer)    
app.register_blueprint(bpChannel)

def run():
    database.initialize()
    app.run(port=8080)

run()