from flask import request, jsonify, Response
from init import app, cors
import database
import AuthKeyGen
from AuthKeyGen import token_required
from werkzeug.security import check_password_hash,generate_password_hash
import datetime

@app.before_request
def before_request(): 
    allowed_origin = request.headers.get('Origin')
    # if allowed_origin in ['http://localhost:4100']:
    cors._origins = allowed_origin

@app.route("/")
def home():
    return "eventually lets put something cool here"

@app.route("/gatekeep",methods=["POST"])
@token_required
def gatekeep(userData):
    return {"userData":userData}

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
        "joindate": datetime.date.today().strftime("%Y-%m-%d")
    }

    database.addRowToTable(createdUser,"user")
    key = AuthKeyGen.encryptJWT({"userID":data["userID"]},60)
    response = jsonify({
            "RelayJWT":key
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
            "RelayJWT":key
        })
    return response

def run():
    database.initialize()
    app.run(port=8080)

run()