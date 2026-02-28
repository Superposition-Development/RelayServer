from dotenv import load_dotenv
from flask import Flask
from flask_cors import CORS
import os
load_dotenv()
SERVER_NAME = os.getenv("SERVER_NAME")
CUSTOM_DATABASE_PATH = os.getenv("CUSTOM_DATABASE_PATH")

if CUSTOM_DATABASE_PATH:
    if not os.path.exists(CUSTOM_DATABASE_PATH):
        os.makedirs(CUSTOM_DATABASE_PATH)
    
    DATABASE_NAME = os.path.join(CUSTOM_DATABASE_PATH, os.getenv("DATABASE_NAME"))
else:
    os_data_path = ""
    match os.name:
        case 'nt':
            os_data_path = os.path.join(os.environ.get("LOCALAPPDATA", ""), "relay")
        case 'posix':
            os_data_path = os.path.join(os.environ.get("XDG_DATA_HOME", ""), "relay")
    
    if not os.path.exists(os_data_path):
        os.makedirs(os_data_path)

    DATABASE_NAME = os.path.join(os_data_path, os.getenv("DATABASE_NAME"))
    



SECRET_KEY = os.getenv("SECRET_KEY")
SIGNUP_PASSWORD_REQUIRED = os.getenv("SIGNUP_PASSWORD_REQUIRED") == "True"
SIGNUP_PASSWORD = os.getenv("SIGNUP_PASSWORD")

app = Flask(SERVER_NAME)
cors = CORS(app=app,supports_credentials=True,origins="*")
