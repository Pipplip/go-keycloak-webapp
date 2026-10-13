Ich möchte ein modernes Lernprojekt entwickeln, das zeigt, wie Authentifizierung und Autorisierung in einer professionellen Go-Webanwendung umgesetzt werden.

Ziele des Projekts:

Die Konzepte hinter OpenID Connect (OIDC), OAuth 2.0, Sessions und rollenbasierter Autorisierung (RBAC) verstehen.
Eine Architektur aufbauen, die sich an Best Practices orientiert und sich auch für produktive Anwendungen eignen würde.
Nicht nur Code schreiben, sondern auch die Hintergründe und Designentscheidungen verstehen.
Tech-Stack
Backend: Go (möglichst nah an der Standardbibliothek)
HTTP: net/http oder chi
Authentifizierung: Keycloak als OIDC Identity Provider
Autorisierung: Rollen aus Keycloak, später optional eigene Policies
Session-Management: Serverseitige Sessions (keine JWT-basierte Browser-Authentifizierung)
Datenbank: PostgreSQL
Containerisierung: Docker Compose
Frontend: Server Side Rendering mit html/template
Deployment: Lokale Entwicklungsumgebung
Projektstruktur

Ich möchte eine saubere, skalierbare Projektstruktur verwenden, z. B.:

project/
├── cmd/
├── internal/
│   ├── auth/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   ├── repository/
│   ├── services/
│   └── templates/
├── migrations/
├── docker-compose.yml
├── Dockerfile
└── README.md
Lernansatz

Bitte arbeite mit mir inkrementell.

Nicht das komplette Projekt auf einmal generieren.
Immer nur den nächsten sinnvollen Schritt.
Erkläre zunächst die Theorie.
Begründe Architekturentscheidungen.
Danach implementieren wir gemeinsam.
Nach jedem Schritt überprüfen wir das Ergebnis, bevor wir weitermachen.
Wichtige Anforderungen

Ich möchte möglichst viel selbst verstehen.

Bitte:

erkläre OIDC, OAuth2 und Keycloak während der Implementierung,
erkläre jede verwendete Bibliothek,
erkläre, warum eine Lösung Best Practice ist,
weise auf Sicherheitsaspekte hin,
erkläre Alternativen und deren Vor- und Nachteile.
Authentifizierung

Ich möchte keine eigene Login-Lösung entwickeln.

Stattdessen soll Keycloak verwendet werden.

Die Go-Anwendung soll als OIDC-Client fungieren und den vollständigen Authorization-Code-Flow implementieren.

Features

Das Projekt soll schrittweise folgende Funktionen erhalten:

Docker Compose mit Keycloak und PostgreSQL
Go-Webserver
OIDC Login
Logout
Serverseitige Sessions
Geschützte Routen
Benutzer beim ersten Login in der Datenbank anlegen
Rollenbasierte Autorisierung
Eigene Benutzerdaten (Profil)
CRUD-Beispiel (z. B. Projekte oder Aufgaben)
Middleware
Logging
Konfiguration über Environment-Variablen
Optional: CSRF-Schutz
Optional: E-Mail-Verifizierung
Optional: MFA
Optional: GitHub- oder Google-Login über Keycloak
Codequalität

Ich möchte sauberen, idiomatischen Go-Code schreiben.

Bitte achte auf:

Clean Architecture, soweit für ein Lernprojekt sinnvoll
Dependency Injection ohne Framework
kleine Pakete mit klarer Verantwortung
gute Fehlerbehandlung
verständliche Kommentare nur dort, wo sie einen Mehrwert bieten
keine unnötigen Frameworks
möglichst Standardbibliothek bevorzugen

Ziel

Am Ende möchte ich nicht nur eine funktionierende Webanwendung besitzen, sondern verstehen,

wie moderne Authentifizierung funktioniert,
warum OIDC heute Standard ist,
wie Keycloak integriert wird,
wie Sessions funktionieren,
wie Autorisierung umgesetzt wird,
und wie professionelle Go-Webanwendungen strukturiert werden.

Bitte agiere dabei wie ein erfahrener Software-Architekt und Go-Entwickler, der Wert auf moderne Best Practices, verständliche Erklärungen und nachhaltiges Lernen legt.