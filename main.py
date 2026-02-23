from flask import request, jsonify
from init import app, cors
import database
from AuthKeyGen import requiresToken
import time
from routes.auth import bpAuth

@app.before_request
def before_request(): 
    allowed_origin = request.headers.get('Origin')
    # if allowed_origin in ['http://localhost:4100']:
    cors._origins = allowed_origin

@app.route("/")
def home():
    return "eventually lets put something cool here"

app.register_blueprint(bpAuth)

def createServerUser(serverID,userID):
    serverUser = {
        "serverID": serverID,
        "userID": userID,
        "timestamp": int(time.time())
    }
    database.addRowToTable(serverUser,"serverUser")
    
"""
expected payload:

serverName: value
pfp: value
userID: value

"""

@app.route("/createServer",methods=["POST"])
@requiresToken
def createServer(user):
    data = request.get_json()

    createdServer = {
        "name": data["name"],
        "pfp": data["pfp"],
        "timestamp": int(time.time())
    }
    serverID = database.addRowToTable(createdServer,"server")
    createServerUser(serverID,user[3])

    response = jsonify({
            "Message":"Server Created Successfully"
        })
    return response

def run():
    database.initialize()
    app.run(port=8080)

run()