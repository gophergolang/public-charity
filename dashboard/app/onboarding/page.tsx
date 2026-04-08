import { redirect } from "next/navigation";
import { getOnboardingStep } from "@/lib/actions/onboarding";

export default async function OnboardingIndex() {
  const step = await getOnboardingStep();
  if (step >= 7) redirect("/dashboard");
  redirect(`/onboarding/step/${Math.max(1, step + 1)}`);
}
