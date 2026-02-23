import jwt
import datetime
import init
from flask import jsonify, request
from functools import wraps
import database

secretKey = init.SECRET_KEY

def encryptJWT(payload,expireTimeMin:int):
    payload["exp"] = datetime.datetime.utcnow() + datetime.timedelta(minutes=expireTimeMin)
    return jwt.encode(payload=payload,key=secretKey,algorithm="HS256")

def decryptJWT(token):
    status = {"Error":"JWT decryption failed for unknown reason"}
    try:
        status = {"Result":jwt.decode(jwt=token,key=secretKey,algorithms=["HS256"])}
    except jwt.ExpiredSignatureError:
        status = {"Error":"JWT expired"}
    except jwt.InvalidSignatureError:
        status = {"Error":"https://tenor.com/view/i-know-gif-3951224689799851379"}
    return status

def requiresToken(f):
    """
    Add this to an endpoint to indicate the following function should require a user to have authenticated with the server to proceed

    A sample of this functionality could be
    ```
    @bpExample.route("/restrictedAction",methods=["POST"])
    @requiresToken
    def restrictedAction(user):
        print(user["userID"]) # get other data by putting another index into user[]
    ```

    Will error if the token is invalid, expects the token to be in the headers by Authorization: Bearing {token}
    
    """
    @wraps(f)
    def decorated(*args, **kwargs):
        token = request.headers.get("Authorization")
        if not token:
            return jsonify({'Error': 'Unauthorized Access, missing JWT'}), 401
        
        if token and token.startswith("Bearer "):
            token = token[7:]

        try:
            result = decryptJWT(token)
            if("Error" in result.keys()):
                return jsonify({'Error': 'Invalid JWT'}), 401
            
            user = database.queryTableValue(["id","pfp","username","userID","password","timestamp"],"user","userID",result["Result"]["userID"])
            if(user == None):
                return jsonify({'Error': 'Invalid JWT'}), 401
            userData = {
                "id":user[0][0],
                "pfp":user[0][1],
                "username":user[0][2],
                "userID":user[0][3],
                "password":user[0][4],
                "timestamp":user[0][5]
            }
    
        except Exception as e:
            return jsonify({'Error': e}), 401

        return f(userData, *args, **kwargs)

    return decorated