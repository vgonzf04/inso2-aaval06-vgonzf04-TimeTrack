"use client"
import * as React from "react"

import { Calendar } from "@/components/ui/calendar"
import { Button } from "@/components/ui/button"
import {
  SidebarGroup,
  SidebarGroupContent,
} from "@/components/ui/sidebar"

export function DatePicker({ onVacacionCreada }) {
  const [selected, setSelected] = React.useState(undefined)
  const [loading, setLoading] = React.useState(false)

  const handleRequestVacation = async () => {
    if (!selected?.from || !selected?.to) return

    setLoading(true)
    try {
      const res = await fetch("http://localhost:3000/vacaciones", {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          inicio: selected.from.toISOString().slice(0, 10),
          fin: selected.to.toISOString().slice(0, 10),
        }),
      })

      if (!res.ok) {
        alert("Error al enviar la solicitud de vacaciones")
        return
      }

      const nuevaVacacion = await res.json()
      console.log("▶ DatePicker recibí del backend:", nuevaVacacion)

      if (onVacacionCreada) {
        onVacacionCreada(nuevaVacacion)
      }

      setSelected(undefined)
    } catch (error) {
      console.error("Error en handleRequestVacation:", error)
      alert("Error al enviar la solicitud de vacaciones")
    } finally {
      setLoading(false)
    }
  }

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
              selected === undefined ||
              selected.from === undefined ||
              selected.to === undefined
            }
          >
            {loading ? "Enviando..." : "Pedir vacaciones"}
          </Button>
        </div>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}
