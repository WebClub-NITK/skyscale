def handler(event, context):
    """
    Advanced Lambda handler example showing API Gateway integration
    
    This function demonstrates how to handle API Gateway events, including:
    - Different HTTP methods (GET, POST, PUT, DELETE)
    - Path parameters
    - Query string parameters
    - Request body parsing
    - Returning structured responses
    
    Parameters:
    event (dict): The event from API Gateway
    context (object): Lambda execution context
    
    Returns:
    dict: API Gateway response object
    """
    try:
        # Log request details
        print(f"Request ID: {context.request_id}")
        print(f"Remaining time: {context.get_remaining_time_in_millis()} ms")
        print(f"Event: {event}")
        
        # Extract API Gateway specific information
        http_method = event.get('httpMethod', '')
        resource_path = event.get('path', '')
        stage = event.get('requestContext', {}).get('stage', '')
        query_params = event.get('queryStringParameters', {}) or {}
        path_params = event.get('pathParameters', {}) or {}
        body = event.get('body', None)
        
        # Initialize the response
        response = {
            "statusCode": 200,
            "headers": {
                "Content-Type": "application/json",
                "Access-Control-Allow-Origin": "*"  # Allow CORS
            },
            "body": {}
        }
        
        # Process based on HTTP method and path
        if http_method == 'GET':
            if '/items' in resource_path:
                # GET /items - List all items
                if not path_params:
                    # Implement pagination if provided
                    limit = int(query_params.get('limit', 10))
                    offset = int(query_params.get('offset', 0))
                    
                    # Mock data - in a real app, you'd fetch from a database
                    items = [{"id": i, "name": f"Item {i}"} for i in range(10)]
                    
                    response["body"] = {
                        "items": items[offset:offset+limit],
                        "pagination": {
                            "total": len(items),
                            "limit": limit,
                            "offset": offset
                        }
                    }
                else:
                    # GET /items/{id} - Get a specific item
                    item_id = path_params.get('id')
                    # Mock data - in a real app, you'd fetch from a database
                    response["body"] = {"id": item_id, "name": f"Item {item_id}"}
        
        elif http_method == 'POST':
            if '/items' in resource_path:
                # POST /items - Create a new item
                # Parse request body
                import json
                item_data = json.loads(body) if body else {}
                
                # Validate request
                if not item_data.get('name'):
                    response["statusCode"] = 400
                    response["body"] = {"error": "Item name is required"}
                else:
                    # Mock create - in a real app, you'd save to a database
                    response["statusCode"] = 201  # Created
                    response["body"] = {
                        "id": "new-id-123",
                        "name": item_data.get('name'),
                        "created": True
                    }
        
        elif http_method == 'PUT':
            if '/items' in resource_path and path_params.get('id'):
                # PUT /items/{id} - Update an item
                item_id = path_params.get('id')
                # Parse request body
                import json
                item_data = json.loads(body) if body else {}
                
                # Mock update - in a real app, you'd update in a database
                response["body"] = {
                    "id": item_id,
                    "name": item_data.get('name', f"Updated Item {item_id}"),
                    "updated": True
                }
        
        elif http_method == 'DELETE':
            if '/items' in resource_path and path_params.get('id'):
                # DELETE /items/{id} - Delete an item
                item_id = path_params.get('id')
                
                # Mock delete - in a real app, you'd delete from a database
                response["body"] = {
                    "id": item_id,
                    "deleted": True
                }
        
        else:
            # Unsupported method or route
            response["statusCode"] = 404
            response["body"] = {"error": "Route not found"}
        
        # Serialize the body to JSON string
        if isinstance(response["body"], dict):
            import json
            response["body"] = json.dumps(response["body"])
        
        # Add execution metadata to response headers
        response["headers"]["X-Request-ID"] = context.request_id
        response["headers"]["X-Function-Name"] = context.function_name
        
        return response
        
    except Exception as e:
        # Handle errors
        import traceback
        print(f"Error: {str(e)}")
        print(traceback.format_exc())
        
        return {
            "statusCode": 500,
            "headers": {
                "Content-Type": "application/json"
            },
            "body": json.dumps({
                "error": str(e),
                "request_id": context.request_id
            })
        } 