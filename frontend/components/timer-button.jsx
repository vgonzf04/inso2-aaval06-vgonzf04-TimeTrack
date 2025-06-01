"use client";
import React from "react";
import { Button } from "@/components/ui/button";
import { IconPlayerPlay } from "@tabler/icons-react";

export function TimerButton({ onFichajeCreado, onFichajeCerrado, ...props }) {
  const [coords, setCoords] = React.useState(null);
  const [currentTimer, setCurrentTimer] = React.useState(null);
  const [dateStart, setDateStart] = React.useState(null);

  React.useEffect(() => {
    // Pedir permisos de geolocalización
    if (!("geolocation" in navigator)) {
      alert("Tu navegador no soporta Geolocalización");
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setCoords({
          lat: position.coords.latitude,
          lon: position.coords.longitude,
        });
      },
      (err) => {
        console.warn("Error pidiendo geolocalización:", err);
      },
      {
        enableHighAccuracy: true,
        timeout: 10_000,
        maximumAge: 0,
      }
    );
  }, []);

  React.useEffect(() => {
    // Al montar, comprobamos si ya hay un fichaje abierto
    fetch("http://localhost:3000/fichajes/current", {
      method: "GET",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
    })
      .then((response) => response.json())
      .then((data) => {
        if (data.running) {
          setCurrentTimer(data);
          setDateStart(new Date(data.dateStart));
        } else {
          setCurrentTimer(null);
          setDateStart(null);
        }
      })
      .catch((error) => {
        console.error("Error fetching timer status:", error);
      });
  }, []);

  const toggleTimer = () => {
    // Si dateStart NO es null, significa que hay un fichaje abierto => lo cerramos
    if (dateStart) {
      fetch(`http://localhost:3000/fichajes/${currentTimer.id}/cerrar`, {
        method: "PUT",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          latitud: coords?.lat,
          longitud: coords?.lon,
        }),
      })
        .then((response) => {
          if (!response.ok) throw new Error("Error al cerrar fichaje");
          return response.json();
        })
        .then((fichajeCerrado) => {
          // 1) Cancelamos el estado interno
          setCurrentTimer(null);
          setDateStart(null);

          // 2) Avisamos al padre de que hemos cerrado un fichaje
          //    (por si quiere actualizar alguna tabla de fichajes “cerrados”)
          onFichajeCerrado && onFichajeCerrado(fichajeCerrado);
        })
        .catch((error) => {
          console.error("Error cerrando fichaje:", error);
          alert("No se pudo cerrar el fichaje");
        });

      return;
    }

    // Si dateStart es null, no hay fichaje abierto => lo creamos
    if (!coords) {
      alert("Activa los servicios de ubicación primero");
      return;
    }
    fetch("http://localhost:3000/fichajes", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        latitud: coords.lat,
        longitud: coords.lon,
      }),
    })
      .then((response) => {
        if (!response.ok) throw new Error("Error al iniciar fichaje");
        return response.json();
      })
      .then((nuevoFichaje) => {
        // 1) Actualizar el estado interno de TimerButton
        setCurrentTimer(nuevoFichaje);
        setDateStart(new Date(nuevoFichaje.dateStart));

        // 2) Avisar al componente padre que se acaba de crear un fichaje nuevo
        onFichajeCreado && onFichajeCreado(nuevoFichaje);
      })
      .catch((error) => {
        console.error("Error iniciando fichaje:", error);
        alert("No se pudo iniciar el fichaje");
      });
  };

  return (
    <div onClick={toggleTimer} className="ml-auto flex items-center gap-2">
      <Button size="sm" className="hidden sm:flex cursor-pointer">
        <IconPlayerPlay className="size-4" /> {dateStart ? "Parar día" : "Comenzar día"}
      </Button>
    </div>
  );
}
