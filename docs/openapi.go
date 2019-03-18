package docs

const doc = `
{
  "components": {
    "schemas": {
      "Login": {
        "properties": {
          "email": {
            "type": "string"
          },
          "password": {
            "type": "string"
          },
          "browserToken": {
            "type": "string"
          }
        }
      },
      "LoginResponse": {
        "properties": {
          "code": {
            "type": "integer"
          },
          "expire": {
            "type": "string"
          },
          "token": {
            "type": "string"
          }
        }
      },
      "LoginFail": {
        "properties": {
          "code": {
            "type": "integer"
          },
          "message": {
            "type": "string"
          }
        }
      }
    },
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer"
      }
    }
  },
  "info": {
    "contact": {
      "name": "contact support",
      "email": "contact@signaux-faibles.beta.gouv.fr"
    },
    "description": "datapi is a REST database manager. More information on https://github.com/  ",
    "license": {
      "name": "Licence MIT",
      "url": "https://raw.githubusercontent.com/entrepreneur-interet-general/opensignauxfaibles/master/LICENSE"
    },
    "title": "API openSignauxFaibles",
    "version": "1.1"
  },
  "paths": {
    "/login": {
      "post": {
        "description": "L'authentification se base sur la fourniture de 3 champs\n- email\n- mot de passe \n- jeton du navigateur\n\nLe service retourne un jeton temporaire valide pour une durée déterminée.  \nCette validité peut être prolongée avec le service /refresh \n\nCe token doit être utilisé dans l'entête ` + "`Authorization`" + `",
        "requestBody": {
          "description": "Les informations d'identification",
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/Login"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "retourne un token et une date de validité",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/LoginResponse"
                },
                "example": {
                  "code": 200,
                  "expire": "2018-12-31 23:30:12",
                  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9…"
                }
              }
            }
          },
          "401": {
            "description": "retourne la raison de l'échec",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/LoginFail"
                },
                "example": {
                  "code": 401,
                  "message": "Please provide good credentials."
                }
              }
            }
          }
        },
        "summary": "obtenir un jeton temporaire de session",
        "tags": [
          "Authentification"
        ]
      }
    },
    "/refresh": {
      "get": {
        "summary": "renouveler le jeton de session",
        "description": "Fournit un nouveau jeton temporaire de session en échange d'un jeton encore valide.  \nCe service demande d'être authentifié mais ne nécessite aucun paramètre.  \nL'authentification est réalisée avec l'entête HTTP Authorization.  \nExemple:  \n ` + "```Authorization:Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9…```" + `",
        "responses": {
          "200": {
            "description": "Renouvellement autorisé",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/LoginResponse"
                },
                "example": {
                  "code": 200,
                  "expire": "2019-02-01 12:00:05",
                  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9…"
                }
              }
            }
          },
          "401": {
            "description": "Renouvellement refusé",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/LoginFail"
                },
                "example": {
                  "code": 401,
                  "message": "cookie token is empty"
                }
              }
            }
          }
        },
        "security": [
          {
            "BearerAuth": []
          }
        ],
        "tags": [
          "Authentification"
        ]
      }
    }
  },
  "openapi": "3.0.0"
}
`
