from flask import request, jsonify
from init import app, cors
import database
import AuthKeyGen
from AuthKeyGen import token_required
from werkzeug.security import check_password_hash

@app.before_request
def before_request(): 
    allowed_origin = request.headers.get('Origin')
    if allowed_origin in ['http://localhost:4100']:
        cors._origins = allowed_origin

@app.route("/")
def home():
    return "eventually lets put something cool here"

#TODO: sanitize userID
@app.route("/signup",methods=["POST"])
def signup():
    data = request.get_json()
    if(database.queryTableValue("id","user","userID",data["userID"])[0] != None):
         return jsonify({"Error":"UserID taken"})
    database.createUser(data["username"],data["password"],data["userID"],data["pfp"])
    return jsonify(AuthKeyGen.encryptJWT(
        {
            "userID":data["userID"]
    },1))

@app.route("/login",methods=["POST"])
def login():
    data = request.get_json()
    userQuery = database.queryTableValue(["password","userID"],"user","userID",data["userID"])

    if(userQuery == None):
        return jsonify({"Error":"Invalid Credentials"}) #sure they could just check with /signup to scan for userIDs, but wtv, hopefully anti brute force is written
    if(not(check_password_hash(userQuery[0],data["password"])) or userQuery[1] != data["userID"]):
        return jsonify({"Error":"Invalid Credentials"})
    return jsonify(AuthKeyGen.encryptJWT(
        {
            "userID":data["userID"]
    },60))

def run():
    database.initialize()
    app.run(port=8080)

run()