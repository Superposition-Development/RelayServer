from flask import Flask, request, jsonify
import init
from init import app, cors
import database
import AuthKeyGen
from AuthKeyGen import token_required

@app.before_request
def before_request(): 
    allowed_origin = request.headers.get('Origin')
    if allowed_origin in ['http://localhost:4100']:
        cors._origins = allowed_origin

@app.route("/")
def home():
    return "eventually lets put something cool here"

@app.route("/test",methods=["POST"])
@token_required
def test():

    data = request.get_json()
    
    return data

#TODO: sanitize userID
@app.route("/signup",methods=["POST"])
def signup():
    data = request.get_json()
    if(database.queryUser(data["userID"]) != None):
         return jsonify({"Error":"UserID taken"})
    database.
    return jsonify(AuthKeyGen.encryptJWT(
        {
            "userID":data["userID"]
    },1))


def run():
    # print(database.testUser())
    app.run(port=8080)

run()