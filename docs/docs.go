// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Yerobal Gustaf Sekeon",
            "email": "yerobalg@gmail.com"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/ping": {
            "get": {
                "description": "Check if the server is running",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Server"
                ],
                "summary": "Health Check",
                "responses": {
                    "200": {
                        "description": "PONG!!",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/v1/auth/login": {
            "post": {
                "description": "Login for user",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User"
                ],
                "summary": "Login",
                "parameters": [
                    {
                        "description": "User login body",
                        "name": "loginBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.UserLoginBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/model.HTTPResponse"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "$ref": "#/definitions/model.UserLoginResponse"
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/model.HTTPResponse"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "type": "string"
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/model.HTTPResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "model.HTTPResponse": {
            "type": "object",
            "properties": {
                "data": {},
                "isSuccess": {
                    "type": "boolean"
                },
                "message": {
                    "$ref": "#/definitions/model.ResponseMessage"
                },
                "metaData": {
                    "$ref": "#/definitions/model.Meta"
                },
                "pagination": {
                    "$ref": "#/definitions/model.PaginationParam"
                }
            }
        },
        "model.Meta": {
            "type": "object",
            "properties": {
                "requestId": {
                    "type": "string"
                },
                "timeElapsed": {
                    "type": "string"
                },
                "timestamp": {
                    "type": "string"
                }
            }
        },
        "model.PaginationParam": {
            "type": "object",
            "properties": {
                "currentElement": {
                    "type": "integer"
                },
                "currentPage": {
                    "type": "integer"
                },
                "limit": {
                    "type": "integer"
                },
                "totalElement": {
                    "type": "integer"
                },
                "totalPage": {
                    "type": "integer"
                }
            }
        },
        "model.ResponseMessage": {
            "type": "object",
            "properties": {
                "description": {
                    "type": "string"
                },
                "title": {
                    "type": "string"
                }
            }
        },
        "model.User": {
            "type": "object",
            "properties": {
                "createdAt": {
                    "type": "integer"
                },
                "createdBy": {
                    "type": "integer"
                },
                "deletedAt": {
                    "type": "string",
                    "example": "2020-12-31T00:00:00Z"
                },
                "deletedBy": {
                    "type": "integer"
                },
                "id": {
                    "type": "integer"
                },
                "isFirstLogin": {
                    "type": "boolean"
                },
                "name": {
                    "type": "string"
                },
                "role": {
                    "$ref": "#/definitions/model.role"
                },
                "updatedAt": {
                    "type": "integer"
                },
                "updatedBy": {
                    "type": "integer"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "model.UserLoginBody": {
            "type": "object",
            "required": [
                "password",
                "username"
            ],
            "properties": {
                "password": {
                    "type": "string",
                    "minLength": 8
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "model.UserLoginResponse": {
            "type": "object",
            "properties": {
                "token": {
                    "type": "string"
                },
                "user": {
                    "$ref": "#/definitions/model.User"
                }
            }
        },
        "model.role": {
            "type": "string",
            "enum": [
                "Admin",
                "Supervisor"
            ],
            "x-enum-varnames": [
                "Admin",
                "Supervisor"
            ]
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "Tigaputera Backend API",
	Description:      "API about financial management for construction company",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
