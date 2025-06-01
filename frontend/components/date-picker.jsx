"use client"
import * as React from "react"

import { Calendar } from "@/components/ui/calendar"
import { Button } from "@/components/ui/button"
import {
  SidebarGroup,
  SidebarGroupContent,
} from "@/components/ui/sidebar"

export function DatePicker() {
  const [selected, setSelected] = React.useState(undefined)

  const handleRequestVacation = () => {
    fetch("http://localhost:3000/vacaciones", {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        inicio: selected.from.toISOString().slice(0, 10), // Format to YYYY-MM-DD
        fin: selected.to.toISOString().slice(0, 10), // Format to YYYY-MM-DD
      }),
    }).then((response) => {
      if (response.ok) {
        alert("Solicitud de vacaciones enviada correctamente")
        setSelected(undefined) // Reset the selected dates
      } else {
        alert("Error al enviar la solicitud de vacaciones")
      }
    }).catch((error) => {
      console.error("Error:", error)
      alert("Error al enviar la solicitud de vacaciones")
    })
  }

  return (
    <SidebarGroup className="px-0">
      <SidebarGroupContent>
        <Calendar mode="range" selected={selected} setSelected={setSelected} className="[&_[role=gridcell].bg-accent]:bg-sidebar-primary [&_[role=gridcell].bg-accent]:text-sidebar-primary-foreground [&_[role=gridcell]]:w-[33px]" />
        <div className="p-3">
          <Button onClick={handleRequestVacation} variant="outline" className="w-full" disabled={selected === undefined || selected.from === undefined || selected.to === undefined}>Pedir vacaciones</Button>
        </div>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}
