// frontend/components/date-picker.jsx
"use client";

import * as React from "react";
import { Calendar } from "@/components/ui/calendar";
import { Button } from "@/components/ui/button";
import {
  SidebarGroup,
  SidebarGroupContent,
} from "@/components/ui/sidebar";

// Helper: formatea una Date local a “YYYY-MM-DD”
function formatLocalYYYYMMDD(date) {
  const yyyy = date.getFullYear();
  const mm = String(date.getMonth() + 1).padStart(2, "0");
  const dd = String(date.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

export function DatePicker({ onVacationCreated }) {
  const [selected, setSelected] = React.useState();
  const [loading, setLoading] = React.useState(false);

  const handleRequestVacation = async () => {
    if (!selected?.from || !selected?.to) return;

    setLoading(true);
    try {
      const startDate = formatLocalYYYYMMDD(selected.from);
      const endDate   = formatLocalYYYYMMDD(selected.to);

      const res = await fetch("http://localhost:3000/vacations", {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ startDate, endDate }),
      });

      if (!res.ok) {
        alert("Error al enviar la solicitud de vacaciones");
        return;
      }

      const nuevaVacacion = await res.json();
      console.log("▶ DatePicker recibió:", nuevaVacacion);

      // Inyecta al vuelo la nueva vacación en la tabla del padre
      onVacationCreated?.(nuevaVacacion);
      setSelected(undefined);
    } catch (error) {
      console.error("Error en handleRequestVacation:", error);
      alert("Error al enviar la solicitud de vacaciones");
    } finally {
      setLoading(false);
    }
  };

  return (
    <SidebarGroup className="px-0">
      <SidebarGroupContent>
        <Calendar
          mode="range"
          selected={selected}
          setSelected={setSelected}
          className="[&_[role=gridcell].bg-accent]:bg-sidebar-primary [&_[role=gridcell].bg-accent]:text-sidebar-primary-foreground [&_[role=gridcell]]:w-[33px]"
        />
        <div className="p-3">
          <Button
            onClick={handleRequestVacation}
            variant="outline"
            className="w-full"
            disabled={
              loading ||
              !selected?.from ||
              !selected?.to
            }
          >
            {loading ? "Enviando..." : "Pedir vacaciones"}
          </Button>
        </div>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}
