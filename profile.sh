#!/bin/bash

# Script de profiling pour l'application Go
# Usage: ./profile.sh [cpu|mem|all|bench]

set -e

TYPE="${1:-all}"
DURATION="${2:-30}"
SERVER_URL="${3:-http://localhost:8080}"

PROFILE_DIR="profiles"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

echo -e "${CYAN}=====================================${NC}"
echo -e "${CYAN}   Profiling de l'application Go${NC}"
echo -e "${CYAN}=====================================${NC}"
echo ""

# Crée le dossier de profils
if [ ! -d "$PROFILE_DIR" ]; then
    mkdir -p "$PROFILE_DIR"
    echo -e "${GREEN}[+] Dossier 'profiles' créé${NC}"
fi

# Fonction pour vérifier si le serveur tourne
check_server() {
    if curl -s -f "$SERVER_URL/api/health" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Fonction pour capturer le profil CPU
capture_cpu_profile() {
    echo -e "\n${YELLOW}[*] Capture du profil CPU (durée: $DURATION secondes)...${NC}"
    echo -e "${YELLOW}[*] Faites des requêtes pendant ce temps pour capturer le profil sous charge${NC}"

    CPU_FILE="$PROFILE_DIR/cpu_$TIMESTAMP.prof"

    if curl -s "$SERVER_URL/debug/pprof/profile?seconds=$DURATION" -o "$CPU_FILE"; then
        echo -e "${GREEN}[+] Profil CPU sauvegardé: $CPU_FILE${NC}"

        # Analyse du profil
        echo -e "${YELLOW}[*] Analyse du profil CPU...${NC}"
        echo ""
        echo -e "${CYAN}Top 10 des fonctions les plus coûteuses:${NC}"
        go tool pprof -top -nodecount=10 "$CPU_FILE"

        echo -e "\n${MAGENTA}[i] Pour une analyse interactive, exécutez:${NC}"
        echo -e "${WHITE}    go tool pprof $CPU_FILE${NC}"
        echo -e "${MAGENTA}[i] Pour visualiser en web UI:${NC}"
        echo -e "${WHITE}    go tool pprof -http=:8081 $CPU_FILE${NC}"
    else
        echo -e "${RED}[-] Erreur lors de la capture du profil CPU${NC}"
    fi
}

# Fonction pour capturer le profil mémoire
capture_mem_profile() {
    echo -e "\n${YELLOW}[*] Capture du profil mémoire...${NC}"

    MEM_FILE="$PROFILE_DIR/mem_$TIMESTAMP.prof"

    if curl -s "$SERVER_URL/debug/pprof/heap" -o "$MEM_FILE"; then
        echo -e "${GREEN}[+] Profil mémoire sauvegardé: $MEM_FILE${NC}"

        # Analyse du profil
        echo -e "${YELLOW}[*] Analyse du profil mémoire...${NC}"
        echo ""
        echo -e "${CYAN}Top 10 des allocations mémoire:${NC}"
        go tool pprof -top -nodecount=10 "$MEM_FILE"

        echo -e "\n${MAGENTA}[i] Pour une analyse interactive, exécutez:${NC}"
        echo -e "${WHITE}    go tool pprof $MEM_FILE${NC}"
        echo -e "${MAGENTA}[i] Pour visualiser en web UI:${NC}"
        echo -e "${WHITE}    go tool pprof -http=:8081 $MEM_FILE${NC}"
    else
        echo -e "${RED}[-] Erreur lors de la capture du profil mémoire${NC}"
    fi
}

# Fonction pour exécuter les benchmarks
run_benchmarks() {
    echo -e "\n${YELLOW}[*] Exécution des benchmarks...${NC}"
    echo ""

    BENCH_FILE="$PROFILE_DIR/bench_$TIMESTAMP.txt"
    BENCH_CPU="$PROFILE_DIR/bench_cpu_$TIMESTAMP.prof"
    BENCH_MEM="$PROFILE_DIR/bench_mem_$TIMESTAMP.prof"

    go test -bench=. -benchmem -cpuprofile="$BENCH_CPU" -memprofile="$BENCH_MEM" | tee "$BENCH_FILE"

    echo ""
    echo -e "${GREEN}[+] Résultats des benchmarks sauvegardés: $BENCH_FILE${NC}"
    echo -e "${GREEN}[+] Profils CPU et mémoire des benchmarks sauvegardés${NC}"

    echo -e "\n${MAGENTA}[i] Pour analyser les profils des benchmarks:${NC}"
    echo -e "${WHITE}    go tool pprof -http=:8081 $BENCH_CPU${NC}"
    echo -e "${WHITE}    go tool pprof -http=:8081 $BENCH_MEM${NC}"
}

# Fonction pour générer de la charge
generate_load() {
    echo -e "\n${YELLOW}[*] Génération de charge pour le profiling...${NC}"
    echo -e "${YELLOW}[*] Envoi de 5 requêtes /api/stats?days=365...${NC}"

    for i in {1..5}; do
        echo "  Requête $i/5..."
        curl -s "$SERVER_URL/api/stats?days=365" > /dev/null 2>&1 || echo "  Erreur sur la requête $i"
    done

    echo -e "${GREEN}[+] Charge générée${NC}"
}

# Vérification du serveur (sauf pour les benchmarks)
if [ "$TYPE" != "bench" ]; then
    echo -e "${YELLOW}[*] Vérification que le serveur est en cours d'exécution...${NC}"
    if ! check_server; then
        echo -e "${RED}[-] ERREUR: Le serveur n'est pas accessible sur $SERVER_URL${NC}"
        echo -e "${YELLOW}[!] Démarrez le serveur avec: go run main.go${NC}"
        exit 1
    fi
    echo -e "${GREEN}[+] Serveur accessible${NC}"
fi

# Exécution selon le type
case "$TYPE" in
    cpu)
        # Lance une requête en arrière-plan pour générer de la charge
        (sleep 2 && curl -s "$SERVER_URL/api/stats?days=365" > /dev/null 2>&1) &
        capture_cpu_profile
        ;;
    mem)
        generate_load
        capture_mem_profile
        ;;
    bench)
        run_benchmarks
        ;;
    all)
        echo -e "${YELLOW}[*] Profiling complet: CPU + Mémoire${NC}"

        # Lance une charge en arrière-plan
        (
            sleep 2
            for i in {1..10}; do
                curl -s "$SERVER_URL/api/stats?days=365" > /dev/null 2>&1 || true
                sleep 0.5
            done
        ) &
        LOAD_PID=$!

        capture_cpu_profile
        wait $LOAD_PID 2>/dev/null || true

        generate_load
        capture_mem_profile
        ;;
    *)
        echo -e "${RED}[-] Type invalide: $TYPE${NC}"
        echo "Usage: $0 [cpu|mem|all|bench] [duration] [server_url]"
        exit 1
        ;;
esac

echo ""
echo -e "${CYAN}=====================================${NC}"
echo -e "${CYAN}   Profiling terminé!${NC}"
echo -e "${CYAN}=====================================${NC}"
echo ""
echo -e "${MAGENTA}[i] Les profils sont disponibles dans le dossier: $PROFILE_DIR${NC}"
echo -e "${MAGENTA}[i] Pour une analyse visuelle interactive:${NC}"
echo -e "${WHITE}    go tool pprof -http=:8081 <fichier.prof>${NC}"