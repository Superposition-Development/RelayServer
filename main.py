from flask import request, jsonify, Response
from init import app, cors
import database
import AuthKeyGen
from AuthKeyGen import requiresToken
from werkzeug.security import check_password_hash,generate_password_hash
import time

@app.before_request
def before_request(): 
    allowed_origin = request.headers.get('Origin')
    # if allowed_origin in ['http://localhost:4100']:
    cors._origins = allowed_origin

@app.route("/")
def home():
    return "eventually lets put something cool here"

#TODO: sanitize userID
@app.route("/signup",methods=["POST"])
def signup():
    data = request.get_json()
    if(database.queryTableValue("id","user","userID",data["userID"]) != None):
         return jsonify({"Error":"UserID taken"})
    
    createdUser = {
        "username": data["username"],
        "password": generate_password_hash(data['password'], method='pbkdf2:sha256'),
        "userID": data["userID"],
        "pfp": data["pfp"],
        "timestamp": int(time.time())
    }

    database.addRowToTable(createdUser,"user")
    key = AuthKeyGen.encryptJWT({"userID":data["userID"]},60)
    response = jsonify({
            "RelayJWT":key,
            "userID":data["userID"]
        })
    return response

@app.route("/login",methods=["POST"])
def login():
    data = request.get_json()
    userQuery = database.queryTableValue(["password","userID"],"user","userID",data["userID"])

    if(userQuery == None):
        return jsonify({"Error":"Invalid Credentials"}) #sure they could just check with /signup to scan for userIDs, but wtv, hopefully anti brute force is written
    if(not(check_password_hash(userQuery[0],data["password"])) or userQuery[1] != data["userID"]):
        return jsonify({"Error":"Invalid Credentials"})
    
    key = AuthKeyGen.encryptJWT({"userID":data["userID"]},60)
    response = jsonify({
            "RelayJWT":key,
            "userID":data["userID"]
        })
    return response


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

# @app.route("/createChannel",methods=["POST"])
# @requiresToken
# def createChannel(user):
#     data = request.get_json()
#     createdServer = {
#         "name": data["name"],
#         "userID": user[3],
#         "pfp": data["pfp"],
#         "timestamp": int(time.time())
#     }
#     database.addRowToTable(createdServer,"server")

#     response = jsonify({
#             "Message":"Server Created Successfully"
#         })
#     return response



def run():
    database.initialize()
    app.run(port=8080)

run()