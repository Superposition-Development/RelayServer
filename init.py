from dotenv import load_dotenv
from flask import Flask
from flask_cors import CORS
import os
load_dotenv()
SERVER_NAME = os.getenv("SERVER_NAME")
DATABASE_NAME = os.getenv("DATABASE_NAME")
SECRET_KEY = os.getenv("SECRET_KEY")
SIGNUP_PASSWORD_REQUIRED = os.getenv("SIGNUP_PASSWORD_REQUIRED") == "True"
SIGNUP_PASSWORD = os.getenv("SIGNUP_PASSWORD")

app = Flask(SERVER_NAME)
cors = CORS(app=app,supports_credentials=True,origins="*")