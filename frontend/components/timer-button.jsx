"use client";
import React from "react";
import { Button } from "@/components/ui/button";
import { IconPlayerPlay } from "@tabler/icons-react";

export function TimerButton({
  onTimecardCreated,
  onTimecardClosed,
  ...props
}) {
  const [coords, setCoords] = React.useState(null);
  const [currentTimer, setCurrentTimer] = React.useState(null);
  const [dateStart, setDateStart] = React.useState(null);

  // 1) Pedir geolocalización
  React.useEffect(() => {
    if (!("geolocation" in navigator)) {
      alert("Tu navegador no soporta Geolocalización");
      return;
    }
    navigator.geolocation.getCurrentPosition(
      ({ coords }) => setCoords({ lat: coords.latitude, lon: coords.longitude }),
      (err) => console.warn("Error pidiendo geolocalización:", err),
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 0 }
    );
  }, []);

  // 2) Al montar, comprobamos si hay un timecard abierto
  React.useEffect(() => {
    fetch("http://localhost:3000/timecards/current", {
      credentials: "include",
    })
      .then((res) => {
        if (res.status === 404) return null; // Ningún timecard abierto
        if (!res.ok) throw new Error("Error fetching current timecard");
        return res.json();
      })
      .then((data) => {
        if (data) {
          setCurrentTimer(data);
          setDateStart(new Date(data.start));
        }
      })
      .catch((err) => console.error("Error fetching timer status:", err));
  }, []);

  const toggleTimer = () => {
    // — Si hay dateStart, entonces cierro el timecard abierto
    if (dateStart) {
      fetch(
        `http://localhost:3000/timecards/${currentTimer.id}/close`,
        {
          method: "PUT",
          credentials: "include",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ lat: coords?.lat, lng: coords?.lon }),
        }
      )
        .then((res) => {
          if (!res.ok) throw new Error("Error al cerrar timecard");
          return res.json();
        })
        .then((closed) => {
          // 1) Actualizar estado interno
          setCurrentTimer(null);
          setDateStart(null);
          // 2) Avisar al padre para que actualice la tabla
          onTimecardClosed?.(closed);
        })
        .catch((err) => {
          console.error("Error cerrando timecard:", err);
          alert("No se pudo cerrar el registro");
        });
      return;
    }

    // — Si no hay dateStart, creamos uno nuevo
    if (!coords) {
      alert("Activa los servicios de ubicación primero");
      return;
    }
    fetch("http://localhost:3000/timecards", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ lat: coords.lat, lng: coords.lon }),
    })
      .then((res) => {
        if (!res.ok) throw new Error("Error al iniciar timecard");
        return res.json();
      })
      .then((newEntry) => {
        // 1) Actualizar estado interno
        setCurrentTimer(newEntry);
        setDateStart(new Date(newEntry.start));
        // 2) Avisar al padre para que inyecte la fila nueva en la tabla
        onTimecardCreated?.(newEntry);
      })
      .catch((err) => {
        console.error("Error iniciando timecard:", err);
        alert("No se pudo iniciar el registro");
      });
  };

  return (
    <div onClick={toggleTimer} className="ml-auto flex items-center gap-2">
      <Button size="sm" className="hidden sm:flex cursor-pointer" {...props}>
        <IconPlayerPlay className="size-4" />{" "}
        {dateStart ? "Parar día" : "Comenzar día"}
      </Button>
    </div>
  );
}
