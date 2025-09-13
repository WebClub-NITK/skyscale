
def handler(event, context):
    """Skyscale function entry point"""
    import requests 
    response = requests.get("https://api.github.com")
    return {"message": "Hello from hello!", "response": response.json()}
