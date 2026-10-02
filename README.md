# 📦 Game Save Backup - Guía de Instalación y Empaquetado

Guía para compilar, empaquetar e instalar **Game Save Backup** en Windows mediante **Inno Setup**.

---

## 📂 1. Preparación de Archivos

Coloca los siguientes 4 archivos en la misma carpeta raíz de trabajo antes de generar el instalador:

```text
Carpeta_Proyecto/
├── guardar-partidas.exe    <-- Ejecutable de Go ya compilado
├── data.json               <-- Configuración de rutas de partidas
├── .env                    <-- Credenciales de GitHub
└── instalador.iss          <-- Script de Inno Setup
```

```
GITHUB_USER=
GITHUB_TOKEN=
GITHUB_REPO=
```

```
[
  {
    "game_id": 1,
    "game_path": "Saved Games\\God of War\\1638",
    "game_name": "God of War 2018"
  }
]
```
