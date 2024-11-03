# GitHub Stats API Go

![GitHub Stats API Go](https://github.com/sup2ak/github-stats-api-go/actions/workflows/release.yml/badge.svg)
![GitHub CI](https://github.com/sup2ak/github-stats-api-go/actions/workflows/ci.yml/badge.svg)


## Aperçu

`github-stats-api-go` est une API basée sur Go qui récupère des statistiques de GitHub pour un utilisateur spécifié. Elle fournit des informations telles que le nombre de followers, de dépôts, d'étoiles et d'organisations associées à l'utilisateur.

## Fonctionnalités

- Récupérer les statistiques d'un utilisateur depuis GitHub
- Inclut des options pour récupérer les étoiles, les followers, les suivis, les dépôts et les organisations
- Serveur HTTP/HTTPS simple pour gérer les requêtes

## Installation

Pour utiliser cette API, vous devez avoir Go installé sur votre machine. Vous pouvez l'installer depuis le [site officiel de Go](https://golang.org/dl/).

1. Initialisez un nouveau module Go :

```bash
go mod init votre_nom_de_projet
```

2. Installez le module en utilisant `go get` :

```bash
go get github.com/SUP2Ak/github-stats-api-go
```

3. Créez un nouveau fichier Go (par exemple, `main.go`) et importez le module :

```go
package main

import (
    "fmt"
    "time"
    githubstats "github.com/sup2ak/github-stats-api-go"
)

func main() {
    // Configuration
    config := githubstats.Config{
        Token: "votre_token_github",
        IP:    "0.0.0.0",
        Port:  "8080",
        Scheme: "http", // ou "https", la valeur nulle est "http"
        Path:  "/stats",
        RateLimit: 10,
        CacheDuration: 1 * time.Hour,
        // et plus encore...
    }
    // Créez une instance de GStats
    stats := githubstats.GStats{}
    // Connectez-vous à l'API GitHub
    err := stats.Connect(config)
    if err != nil {
        log.Fatalf("Erreur de connexion : %v", err)
    }
}
```

4. Exécutez le programme :

```bash
go run main.go
```

## Exemple de requête

Pour obtenir des statistiques pour l'utilisateur `sup2ak`, vous pouvez faire une requête GET :

```bash
curl "http://localhost:8080/stats?username=sup2ak"
```

## Licence

Ce projet est sous la licence [GPL-3.0](LICENSE).

## Contribution

Les contributions sont les bienvenues ! Veuillez ouvrir une issue ou soumettre une pull request pour toute amélioration ou correction de bug.

