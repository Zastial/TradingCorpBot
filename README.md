# TradingCorpBot

Bot Discord réécrit en Go.

## Pré-requis

- Go 1.19 ou plus
- un bot Discord déjà créé
- un compte n8n avec le webhook de réception

## Variables d'environnement

Crée un fichier `.env` à la racine avec au minimum:

- `DISCORD_BOT_TOKEN`
- `DISCORD_APP_ID`
- `N8N_API_KEY`
- `N8N_WEBHOOK_PATH`
- `N8N_BASE_URL` par défaut `https://n8n.zastial.com`
- `PROD` ou `N8N_PROD`

Règle pour n8n:

- si `PROD=true` ou `N8N_PROD=true`, l'URL utilise `/webhook/`
- sinon, elle utilise `/webhook-test/`

Exemple:

```env
DISCORD_BOT_TOKEN=xxx
DISCORD_APP_ID=xxx
N8N_API_KEY=xxx
N8N_WEBHOOK_PATH=b1d0485f-b63e-4f83-942d-8688089bba1c
N8N_BASE_URL=https://n8n.zastial.com
PROD=false
```

## Secrets Docker

Pour un déploiement Swarm sur une VM, crée les secrets une seule fois sur le manager.
Ils ne doivent pas être poussés par la CI.

Exemples:

```bash
printf '%s' 'xxx' | docker secret create DISCORD_BOT_TOKEN -
printf '%s' 'xxx' | docker secret create DISCORD_APP_ID -
printf '%s' 'xxx' | docker secret create N8N_API_KEY -
printf '%s' 'xxx' | docker secret create N8N_WEBHOOK_PATH -
```

Si un secret doit être changé, il faut généralement le supprimer puis le recréer, puis redéployer le stack.

Exemple de déploiement:

```bash
docker stack deploy -c docker-stack.yml trading_corp_bot
```

La CI ne fait ensuite que pousser l'image `zastial/trading-corp-bot:latest` et `zastial/trading-corp-bot:<sha>`.
Le bot lit d'abord les variables d'environnement, puis les secrets montés dans `/run/secrets/<NOM_SECRET>` si la variable n'existe pas.

## Lancement local

1. Installer les dépendances Go avec `go mod tidy`.
2. Lancer le bot avec `go run .`.
3. Vérifier que le bot s'affiche en ligne dans Discord.

## Commandes du bot

### `/analyse ticker:<symbole>`

- valide le ticker
- envoie la requête à n8n
- affiche un message de confirmation ou d'erreur

Exemple:

```text
/analyse ticker:AAPL
```

### `/tickers prefix:<texte>`

- filtre les tickers qui commencent par le texte donné
- renvoie les résultats sous forme de messages lisibles avec emojis

Exemple:

```text
/tickers prefix:aa
```

### `/tickers company:<nom>`

- cherche dans le nom de l'entreprise
- renvoie uniquement les 10 premiers résultats

Exemple:

```text
/tickers company:apple
```

## Structure du projet

- [main.go](main.go) gère le démarrage du bot
- [config.go](config.go) charge la configuration et les variables d'environnement
- [commands.go](commands.go) déclare et traite les commandes Discord
- [nasdaq.go](nasdaq.go) récupère et filtre les tickers Nasdaq
- [n8n.go](n8n.go) envoie la commande `analyse` vers n8n
