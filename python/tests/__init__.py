import chaparral
import os

token = os.getenv("CHAPARRAL_TOKEN")
host = "127.0.0.1"
port = "8080"

if token == "":
    print("warning: token isn't set")

client = chaparral.Client(host, port, token=token)
print(client.get_version("new-object"))
client.close()
