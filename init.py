from dotenv import load_dotenv
from flask import Flask
from flask_cors import CORS
import os
load_dotenv()
SERVER_NAME = "Relay Server"
DATABASE_NAME = os.getenv("DATABASE_NAME")
JWT_KEY = os.getenv("JWT_KEY")

app = Flask(SERVER_NAME)
cors = CORS(app=app,supports_credentials=True)