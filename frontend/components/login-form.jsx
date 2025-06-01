import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
export function LoginForm({
  className,
  ...props
}) {
  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader>
          <CardTitle>Login to your account</CardTitle>
          <CardDescription>
            Enter your email below to login to your account
          </CardDescription>
        </CardHeader>
        <CardContent>
            <div className="flex flex-col gap-6">
              <div className="flex flex-col gap-3">
                <a href="http://localhost:3000/auth/google"><Button variant="outline" className="w-full">
                  Login with Google
                </Button></a>
              </div>
            </div>
        </CardContent>
      </Card>
    </div>
  )
}