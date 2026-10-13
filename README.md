# Go-Keycloak-Webanwendung

Dieses Lernprojekt zeigt, wie eine Go-Webanwendung Benutzer über OpenID
Connect (OIDC) bei Keycloak anmeldet und anschließend serverseitige Sessions
und rollenbasierte Autorisierung verwendet.

## Architektur

```text
                ┌──────────────────┐
                │     Browser      │
                └────────┬─────────┘
                         │ HTTP
                ┌────────▼─────────┐
                │      Go-App      │
                │ localhost:8080   │
                └─────┬───────┬────┘
                      │       │
         ┌────────────┘       └─────────────┐
         │                                   │
┌────────▼─────────┐                ┌───────▼────────┐
│   PostgreSQL     │                │    Keycloak    │
│ localhost:5432   │                │ localhost:8081 │
│                  │                │                │
│ DB: app          │                │ DB: keycloak   │
└──────────────────┘                └────────────────┘
```

- Keycloak verwaltet Identitäten und Anmeldedaten. Die Go-Anwendung implementiert kein eigenes Passwort-Login.
- PostgreSQL wird als Compose-Service `postgres` betrieben. Darin liegen zwei getrennte Datenbanken: `keycloak` für Keycloak und `app` für Anwendungsdaten.
- Innerhalb des Compose-Netzwerks erreicht die Go-App PostgreSQL unter `postgres:5432`. Ein direkt auf dem Host gestarteter Go-Prozess verwendet dagegen `localhost:5432`.
- Anwendungsmigrationen liegen unter `migrations/app`, sind in die Anwendung eingebettet und werden beim Start einmalig angewendet. Keycloaks eigene Datenbank bleibt davon getrennt.

Beim erfolgreichen OIDC-Login wird der lokale Datensatz in `app_users` anhand der eindeutigen Kombination aus OIDC-Issuer und Subject (`sub`) angelegt oder aktualisiert. Die E-Mail-Adresse ist änderbares Profildatum und kein Identitätsschlüssel.

## Projektstruktur

- `cmd/main.go` lädt die Konfiguration, initialisiert Datenbank und OIDC und verdrahtet die Abhängigkeiten.
- `internal/web` registriert Routen und ordnet Handler sowie Middleware zusammen.
- `internal/handlers` enthält die HTTP-Handler für Anmeldung und Profil.
- `internal/middleware` enthält Prüfungen für Session-Authentifizierung, Rollen und CSRF.
- `internal/users` kapselt den Zugriff auf lokale Benutzerdaten.
- `internal/database` initialisiert die PostgreSQL-Verbindung.
- `migrations/app` enthält die Migrationen der Anwendung; `migrations/keycloak` enthält die Datenbankinitialisierung für Keycloak.

## Anwendung starten und stoppen

```bash
make start
```

Die Anwendung ist unter `http://localhost:8080` und die Keycloak-Verwaltung
unter `http://localhost:8081` erreichbar. Der PostgreSQL-Container muss bereit
und der in `OIDC_ISSUER` angegebene Realm in Keycloak angelegt sein, damit die
Go-App erfolgreich startet.

Die Startseite `/` zeigt derzeit nur einen Platzhalter. Der geschützte
Einstieg ist `http://localhost:8080/profile`; alternativ kann der Login direkt
über `http://localhost:8080/login` gestartet werden.

```bash
make stop
```

Beim Stoppen bleiben die Datenbank-Volumes erhalten. `docker compose down -v`
würde sie löschen und sollte nur verwendet werden, wenn die lokalen Daten
absichtlich zurückgesetzt werden sollen.

## Konfiguration

Die lokale Compose-Umgebung liest OIDC-Konfiguration aus `.env`. Diese Datei
wird nicht ins Repository eingecheckt. PostgreSQL wird für den App-Container
mit `APP_DATABASE_URL` auf `postgres:5432` konfiguriert.

| Variable | Bedeutung |
| --- | --- |
| `APP_PORT` | HTTP-Port der Go-App; Standardwert `8080` |
| `APP_DATABASE_URL` | PostgreSQL-Verbindungszeichenfolge. Compose setzt den Container-Host; beim direkten Start auf dem Host `localhost` verwenden. |
| `OIDC_ISSUER` | Öffentlicher OIDC-Issuer; muss exakt dem `iss`-Wert im ID-Token entsprechen |
| `OIDC_DISCOVERY_URL` | Interne URL für OIDC-Discovery und Back-Channel-Endpunkte. Optional; ohne Angabe wird `OIDC_ISSUER` verwendet. |
| `OIDC_CLIENT_ID` | Client-ID der Go-Anwendung in Keycloak |
| `OIDC_CLIENT_SECRET` | Client-Secret dieses vertraulichen OIDC-Clients |
| `OIDC_REDIRECT_URL` | Callback-URL, normalerweise `http://localhost:8080/auth/callback` |
| `OIDC_AUTHORIZATION_URL` | Optionaler, für den Browser erreichbarer Authorization-Endpunkt. Nützlich, wenn OIDC-Discovery intern `keycloak:8080` liefert. |
| `APP_SESSION_TTL` | Serverseitige Session-Lebensdauer; Standardwert `8h` |
| `APP_SESSION_COOKIE_SECURE` | Setzt das Cookie-Attribut `Secure`; für HTTPS auf `true` setzen. Standardmäßig `false` für lokale HTTP-Entwicklung. |

Ein Compose-Hostname wie `keycloak` ist normalerweise nur innerhalb des
Docker-Netzwerks auflösbar. Deshalb kann die App intern
`OIDC_DISCOVERY_URL=http://keycloak:8080/realms/go-realm` verwenden, während
`OIDC_ISSUER=http://localhost:8081/realms/go-realm` den öffentlichen Issuer
festlegt und `OIDC_AUTHORIZATION_URL` auf den vom Browser erreichbaren
Authorization-Endpunkt zeigt. Die App prüft den `iss`-Wert des ID-Tokens
weiterhin exakt gegen `OIDC_ISSUER`; die Issuer-Prüfung wird nicht deaktiviert.

Die Compose-Datei importiert `realm-export.json` derzeit nicht automatisch.
Realm, Client, Benutzer und Rollen müssen daher weiterhin in Keycloak
eingerichtet werden. Wenn der Export als versionierte Vorlage dienen soll,
exportiere ihn nach Konfigurationsänderungen erneut und lasse Benutzer sowie
Secrets aus der Datei heraus.

Die Compose-Datei enthält außerdem noch feste Zugangsdaten für die lokale
Entwicklungsumgebung. Sie dürfen nicht für gemeinsam erreichbare oder
produktive Systeme verwendet werden; vor einer Veröffentlichung sollten sie
durch lokale Umgebungsvariablen oder geeignete Deployment-Secrets ersetzt
werden.

## Keycloak einrichten

Keycloak übernimmt unter anderem Passwortverwaltung, MFA, E-Mail-Verifizierung
und externe Identitätsanbieter. Für dieses Projekt müssen Realm, Client,
Testbenutzer und Rollen einmalig eingerichtet werden:

1. Einen Realm erstellen, dessen Name dem Realm-Pfad in `OIDC_ISSUER` entspricht.
2. Einen vertraulichen OIDC-Client mit aktiviertem Authorization-Code-Flow erstellen. Seine Client-ID muss `OIDC_CLIENT_ID` entsprechen.
3. Als gültige Redirect-URI den Wert aus `OIDC_REDIRECT_URL` eintragen.
4. Einen Testbenutzer anlegen und eine Client-Rolle `admin` für den OIDC-Client erstellen.
5. Die Rolle `admin` dem Testbenutzer zuweisen.
6. Den Rollen-Protocol-Mapper so konfigurieren, dass Client-Rollen im ID-Token enthalten sind.

Keycloak muss vor dem ersten Start der Go-App eingerichtet sein. Nach dem
Anlegen des Realms kann die App mit `make start` erneut gestartet werden.

## OIDC-Anmeldeablauf

OIDC ergänzt OAuth 2.0 um eine standardisierte Identitätsinformation. Die
Go-Anwendung ist der OIDC-Client, Keycloak der Identitätsanbieter, und der
Browser vermittelt die Weiterleitungen. Die App erhält keine Passwörter.

1. Der Browser ruft `GET /login` auf. Die Go-App erzeugt zufällige Werte für `state`, `nonce` und den PKCE-Verifier. Ein kurzlebiges, HttpOnly-Browser-Cookie bindet die Transaktion an denselben Browser.
2. Die App leitet den Browser zu Keycloak weiter. Die Anfrage enthält unter anderem `response_type=code`, `scope=openid`, `state`, `nonce` und den PKCE-Challenge.
3. Nach erfolgreicher Anmeldung leitet Keycloak den Browser an `/auth/callback` zurück. Die URL enthält einen kurzlebigen Authorization Code, aber keine Benutzerinformationen.
4. Die App prüft den browsergebundenen und einmalig verwendbaren `state`-Wert. Anschließend tauscht sie den Code serverseitig zusammen mit dem PKCE-Verifier gegen Tokens ein.
5. Die OIDC-Bibliothek prüft Signatur, Audience und Ablaufzeit des ID-Tokens. Die Anwendung vergleicht den Issuer exakt mit `OIDC_ISSUER` und prüft zusätzlich den enthaltenen Nonce gegen den Wert aus der Login-Transaktion.
6. Die App legt den Benutzer anhand von Issuer und `sub` in PostgreSQL an oder aktualisiert ihn. Danach erzeugt sie eine serverseitige Session und sendet dem Browser nur deren zufällige ID als Cookie.

Das Zugriffstoken (Access-Token) wird derzeit nicht für Aufrufe einer weiteren API verwendet.
Die Browser-Authentifizierung erfolgt über die serverseitige Session, nicht
über ein JWT-Cookie.

## Sessions und CSRF-Schutz

Das Session-Cookie enthält nur eine zufällige ID und ist `HttpOnly` sowie
`SameSite=Lax`. Die zugehörigen Benutzerdaten und Rollen bleiben serverseitig.
Das Cookie läuft passend zur konfigurierten Session-Lebensdauer ab.

Der Session-Store liegt derzeit im Arbeitsspeicher: Ein Neustart beendet daher
alle aktiven Sessions, auch wenn die Benutzerprofile in PostgreSQL erhalten
bleiben. Eine persistente Session-Speicherung ist ein eigener, noch offener
Ausbauschritt.

Logout ist eine zustandsändernde `POST`-Anfrage. Das Profilformular sendet
dafür ein sitzungsgebundenes CSRF-Token mit. Die Middleware prüft das Token,
bevor sie die Session löscht. Der Logout beendet aktuell die lokale App-Session,
nicht die Keycloak-SSO-Session; ein erneuter Login kann daher direkt wieder bei
Keycloak angemeldet sein.

## Rollenbasierte Autorisierung

Keycloak-Rollen müssen per Protocol-Mapper im ID-Token stehen. Die App liest
Realm-Rollen aus `realm_access.roles` und Rollen des konfigurierten Clients
aus `resource_access.<client-id>.roles`. Rollen erhalten getrennte Namen:

- Realm-Rolle: `realm:<rolle>`
- Client-Rolle: `client:<client-id>:<rolle>`

Die Beispielroute `GET /admin` verlangt exakt
`client:<client-id>:admin`. Eine gleichnamige Realm-Rolle oder eine Rolle
eines anderen Clients gewährt keinen Zugriff. Nicht angemeldete Besucher
werden zum Login geleitet; angemeldete Benutzer ohne die benötigte Rolle
erhalten `403 Forbidden`.

Rollen werden beim Login in die Session übernommen. Änderungen oder
Rollenentzüge in Keycloak wirken sich bei bestehenden Sessions daher erst
nach einer erneuten Anmeldung oder dem Ablauf der Session aus.