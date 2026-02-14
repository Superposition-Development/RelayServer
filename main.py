from flask import Flask
import sqlite3
import init
import crypt

app = Flask(init.SERVER_NAME)
connection = sqlite3.connect(f"{init.DATABASE_NAME}.db")

@app.route("/")
def home():
    return "eventually lets put something cool here"

app.run()