import { getAvailabilityList } from "@/lib/actions/availability";
import AvailabilityGrid from "@/components/dashboard/AvailabilityGrid";

export default async function AvailabilityPage() {
  const slots = await getAvailabilityList();

  return (
    <main>
      <h1>When could you meet someone?</h1>
      <p className="muted" style={{ marginBottom: "1.5rem", maxWidth: "30rem" }}>
        Tap the times you&rsquo;re generally free. We&rsquo;ll use this to
        suggest meetups with people nearby.
      </p>
      <AvailabilityGrid initial={slots} />
    </main>
  );
}
