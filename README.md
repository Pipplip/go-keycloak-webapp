# go-keycloak-webapp
Modern webapp in golang to showcasing OICD (OpenID Connect) authentication with Keycloak

# Aufbau
```
                ┌──────────────────┐
                │     Browser      │
                └────────┬─────────┘
                         │
                         │ HTTP
                         │
                ┌────────▼─────────┐
                │      Go App      │
                │    localhost:8080│
                └────────┬─────────┘
                         │
        ┌────────────────┴──────────────┐
        │                               │
        │                               │
┌───────▼────────┐             ┌────────▼────────┐
│   PostgreSQL   │             │    Keycloak     │
│ localhost:5432 │             │ localhost:8081  │
└────────────────┘             └─────────────────┘
```

- Die Go-Anwendung authentifiziert Benutzer nie selbst.
- Die Go-Anwendung spricht später mit Keycloak.
- Die Go-Anwendung speichert eigene Daten in PostgreSQL.

## Anwendung starten

```bash
make start
```

Anwenung stoppen:  
```bash
make stop
```

## URLs

Go-App:  
```
http://localhost:8080
```

Keycloak:  
```
http://localhost:8081
```

Login:
```
http://localhost:8080/login
```

## Keycloak

Keycloak übernimmt die komplette Identitätsverwaltung.

Unsere Go-Anwendung muss sich dadurch nie um folgende Themen kümmern:
- Benutzername/Passwort
- Passwort-Reset
- Passwort-Richtlinien
- MFA
- E-Mail-Verifizierung
- Social Login
- LDAP/Active Directory
- Identitätsnachweise

Was tun in keycloak:  
1. Realm erstellen
2. Client registrieren
3. Testbenutzer anlegen
4. Rollen definieren
5. Redirect-URI konfigurieren
6. OIDC-Discovery verstehen und testen

## OpenID Connect Authorization Code Flow

Bsp: Ein Benutzer besucht http://localhost:8080/profile  

Anwendung erkennt: Benutzer noch nicht angemeldet

```
                       OIDC Flow

┌─────────┐                              ┌───────────┐
│ Browser │                              │ Keycloak  │
└────┬────┘                              └─────┬─────┘
     │                                         │
     │ GET /login                              │
     ▼                                         │
┌──────────┐                                   │
│ Go App   │                                   │
└────┬─────┘                                   │
     │                                         │
     │ Authorization Request                   │
     │ client_id                               │
     │ redirect_uri                            │
     │ state                                   │
     ├────────────────────────────────────────►│
     │                                         │
     │                 Login                   │
     │◄────────────────────────────────────────┤
     │                                         │
     │ Authorization Code                     │
     │◄────────────────────────────────────────┤
     │                                         │
     │ Token Request                           │
     │ code + client_secret                    │
     ├────────────────────────────────────────►│
     │                                         │
     │ ID Token + Access Token                 │
     │◄────────────────────────────────────────┤
     │                                         │
     │                                         │
     │ Session erstellen                       │
     │                                         │
     │ Set-Cookie session_id                  │
     ▼                                         │
┌─────────┐                                   │
│ Browser │                                   │
└─────────┘                                   │
```

In Keycloak ist der Resource Owner der Benutzer, der sich anmeldet.  
Der Client ist unsere App.  
Der Identitiy Provider ist keycloak. Er kennt:  
- Benutzer
- Passwörter
- MFA
- Rollen

### Ablauf: Auth request

1. Login:  
`https://localhost:8081/realms/learning/protocol/openid-connect/auth`

Parameter:
- client_id=go-web-app  
- response_type=code // auth code, keine Tokens
- scope=openid // macht aus OAuth2 → OpenID Connect.
- redirect_uri=http://localhost:8080/auth/callback
- state=xyz123 // wird von der App erstellt

2. Benutzer meldet sich bei Keycloak an. App sieht davon nichts.
3. Nach erfolgreichem Login sendet Keycloak zurück an App
`http://localhost:8080/auth/callback?code=abc123&state=a8f91c`
Wir bekommen nur einen Code, keine Benutzerinfos.

4. Code exchange.  
Unsere App macht einen Server zu Server Aufruf an Keycloack mit code + client_secret   
Keycloak prüft ob Code gültig, zum Client gehört und das Secret korrekt ist.  
Die Antwort ist:  
- access_token - Was darf der Benutzer
- id_token  - Wer ist der Benutzer
- refresh_token - um neue Tokens zu generieren

5. Session erzeugen:  
6. 