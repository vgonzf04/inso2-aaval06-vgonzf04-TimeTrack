"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import EmployeeForm from "@/components/employee";

export default function CreateEmployeePage() {
  const [role, setRole] = useState(null);
  const router = useRouter();

  useEffect(() => {
    async function fetchRole() {
      try {
        const res = await fetch("http://localhost:3000/me", {
          method: "GET",
          credentials: "include",
        });
        if (!res.ok) {
          router.push("/login");
          return;
        }
        const data = await res.json();
        setRole(data.role);
        if (data.role !== "supervisor") {
          router.push("/admin");
        }
      } catch {
        router.push("/login");
      }
    }
    fetchRole();
  }, [router]);

  if (role === null) {
    return <p>Checking permissions...</p>;
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-4">Add New Employee</h1>
      <EmployeeForm />
    </div>
  );
}
