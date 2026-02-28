from dotenv import load_dotenv
from flask import Flask
from flask_cors import CORS
import os
load_dotenv()
SERVER_NAME = os.getenv("SERVER_NAME")
USING_CUSTOM_DB_PATH = os.getenv("USING_CUSTOM_DB_PATH") == "True"

DATABASE_NAME = os.getenv("DATABASE_NAME")

if USING_CUSTOM_DB_PATH:  
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
