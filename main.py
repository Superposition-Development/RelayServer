from flask import Flask, request, jsonify
import sqlite3
import init
from init import app, cors
import crypt

connection = sqlite3.connect(f"{init.DATABASE_NAME}.db")

@app.before_request
def before_request(): 
    allowed_origin = request.headers.get('Origin')
    if allowed_origin in ['http://localhost:4100']:
        cors._origins = allowed_origin

@app.route("/")
def home():
    return "eventually lets put something cool here"

@app.route("/test",methods=["POST"])
def test():

    data = request.get_json()
    
    return data

app.run(port=8080)