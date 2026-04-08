import { redirect, notFound } from "next/navigation";
import { getOnboardingStep } from "@/lib/actions/onboarding";
import { getProfile } from "@/lib/actions/profile";
import { getNeedScores } from "@/lib/actions/needs";
import { getInterests } from "@/lib/actions/interests";
import { getOffers } from "@/lib/actions/offers";
import { getAvailabilityList } from "@/lib/actions/availability";
import StepBasics from "@/components/onboarding/StepBasics";
import StepLocation from "@/components/onboarding/StepLocation";
import StepNeeds from "@/components/onboarding/StepNeeds";
import StepInterests from "@/components/onboarding/StepInterests";
import StepOffers from "@/components/onboarding/StepOffers";
import AvailabilityGrid from "@/components/dashboard/AvailabilityGrid";

const STEP_TITLES = [
  "",
  "Welcome",
  "Your Location",
  "Your Needs",
  "Your Interests",
  "What You Can Offer",
  "Your Week",
];

export default async function OnboardingStepPage({
  params,
}: {
  params: Promise<{ step: string }>;
}) {
  const { step: stepStr } = await params;
  const step = parseInt(stepStr, 10);
  if (isNaN(step) || step < 1 || step > 6) notFound();

  const currentStep = await getOnboardingStep();
  if (currentStep >= 7) redirect("/dashboard");

  return (
    <main>
      <div className="progress-bar">
        <div className="progress-fill" style={{ width: `${(step / 6) * 100}%` }} />
      </div>
      <p className="muted step-label">
        Step {step} of 6 — {STEP_TITLES[step]}
      </p>

      {step === 1 && <StepBasics initial={await getProfile()} />}
      {step === 2 && <StepLocation initial={await getProfile()} />}
      {step === 3 && <StepNeeds initial={await getNeedScores()} />}
      {step === 4 && <StepInterests initial={await getInterests()} />}
      {step === 5 && <StepOffers initial={await getOffers()} />}
      {step === 6 && <AvailabilityGrid initial={await getAvailabilityList()} />}
    </main>
  );
}
