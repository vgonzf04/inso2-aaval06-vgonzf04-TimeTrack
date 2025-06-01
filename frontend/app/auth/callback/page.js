"use client"

import { useEffect } from "react"
import { useSearchParams, useRouter } from "next/navigation"

export default function GoogleCallbackPage() {
  const searchParams = useSearchParams()
  const router = useRouter()

  useEffect(() => {
    const code = searchParams.get("code")

    if (code) {
      fetch(`http://localhost:3000/auth/google/callback?code=${code}`)
        .then(res => res.json())
        .then(data => {

          const rol = data.usuario?.rol
          if (rol) {
            localStorage.setItem("rol", rol)
            router.push("/admin")
            //console.log(">> [Callback Google] respuesta completa:", data)

          } else {
            console.error("No se encontró el rol en la respuesta:", data)
          }
        })
        .catch(err => {
          console.error("Error en callback de Google:", err)
        })
    }
  }, [searchParams, router])

  return <p>Autenticando con Google...</p>
}
