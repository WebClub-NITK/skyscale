def handler(event, context):
    """
    Sample Lambda-style handler function
    
    This function demonstrates the Lambda-style handler pattern with event and context
    
    Parameters:
    event (dict): The event data passed to the function
    context (object): Context object with function execution information
    
    Returns:
    dict: Response containing the processed event data
    """
    # Print context information
    print(f"Function name: {context.function_name}")
    print(f"Request ID: {context.request_id}")
    print(f"Remaining time: {context.get_remaining_time_in_millis()} ms")
    
    # Process the event data
    processed_data = {
        "message": "Hello from FaaS Lambda-style function!",
        "received_event": event,
        "context_info": {
            "function_name": context.function_name,
            "request_id": context.request_id,
            "memory_limit": context.memory_limit_mb,
        }
    }
    
    # You can either return a dict (it will be JSON-serialized)
    # or return a string
    return processed_data 