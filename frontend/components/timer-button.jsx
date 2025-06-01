"use client";
import React from "react";

import { Button } from "@/components/ui/button";
import { IconPlayerPlay } from "@tabler/icons-react";
import { set } from "date-fns";


export function TimerButton({ ...props }) {
    const [coords, setCoords] = React.useState(null);
    const [currentTimer, setCurrentTimer] = React.useState(null);
    const [dateStart, setDateStart] = React.useState(null);
    const [loading, setLoading] = React.useState(true);

    React.useEffect(() => {
        // Make sure the browser supports Geolocation
        if (!('geolocation' in navigator)) {
            setError('Geolocation is not supported by your browser.');
            setLoading(false);
            return;
        }

        // Request the current position
        navigator.geolocation.getCurrentPosition(
            (position) => {
                setCoords({
                    lat: position.coords.latitude,
                    lon: position.coords.longitude,
                });
                setLoading(false);
            },
            (err) => {
                // Possible errors: PERMISSION_DENIED, POSITION_UNAVAILABLE, TIMEOUT
                switch (err.code) {
                    case err.PERMISSION_DENIED:
                        setError('Permission to access location was denied.');
                        break;
                    case err.POSITION_UNAVAILABLE:
                        setError('Position unavailable.');
                        break;
                    case err.TIMEOUT:
                        setError('Location request timed out.');
                        break;
                    default:
                        setError('An unknown error occurred.');
                }
                setLoading(false);
            },
            {
                enableHighAccuracy: true,
                timeout: 10_000, // 10 seconds
                maximumAge: 0,
            }
        );
    }, []);

    React.useEffect(() => {
        // Check if the timer is already running when the component mounts
        fetch("http://localhost:3000/fichajes/current", {
            method: "GET",
            credentials: "include",
            headers: {
                "Content-Type": "application/json",
            }
        })
            .then(response => response.json())
            .then(data => {
                console.log("Timer status:", data);
                if (data.running) {
                    setCurrentTimer(data);
                    setDateStart(new Date(data.dateStart));
                } else {
                    setCurrentTimer(null);
                    setDateStart(null);
                }
            })
            .catch(error => {
                console.error("Error fetching timer status:", error);
            });
    }, []);

    const toggleTimer = () => {
        if (dateStart !== null && dateStart !== undefined) {
            // If the timer is running, stop it
            fetch(`http://localhost:3000/fichajes/${currentTimer.id}/cerrar`, {
                method: "PUT",
                credentials: "include",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    latitud: coords.lat,
                    longitud: coords.lon,
                }),
            })
                .then(response => {
                    if (response.ok) {
                        alert("Timer stopped successfully");
                        setDateStart(null);
                    } else {
                        alert("Error stopping timer");
                    }
                })
                .catch(error => {
                    console.error("Error:", error);
                    alert("Error stopping timer");
                });
            return;
        } else {
            // If the timer is not running, start it
            if (coords === null) {
                alert("Location not available. Please enable location services.");
                return;
            }
            const data = {
                latitud: coords.lat,
                longitud: coords.lon,
            };
            fetch("http://localhost:3000/fichajes", {
                method: "POST",
                credentials: "include",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(data),
            })
                .then(response => {
                    if (!response.ok) {
                        console.error("Error starting timer");
                    }
                    return response.json();
                }).then(data => {
                    console.log("Timer started:", data);
                    setDateStart(new Date(data.dateStart));
                })
                .catch(error => {
                    console.error("Error:", error);
                });
            return;

        }
    };
    return (
        <div onClick={toggleTimer} className="ml-auto flex items-center gap-2">
            <Button size="sm" className="hidden sm:flex cursor-pointer">
                <IconPlayerPlay className="size-4" /> Comenzar el día
            </Button>
        </div>
    )
}