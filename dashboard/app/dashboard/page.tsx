import { getProfile } from "@/lib/actions/profile";
import { getNeedScores } from "@/lib/actions/needs";
import { getInterests } from "@/lib/actions/interests";
import { unreadCount } from "@/lib/actions/inbox";
import NeedSliders from "@/components/dashboard/NeedSliders";
import ProfileBasics from "@/components/dashboard/ProfileBasics";
import InterestEditor from "@/components/dashboard/InterestEditor";
import Link from "next/link";

export default async function DashboardHome() {
  const [profile, scores, interests, unread] = await Promise.all([
    getProfile(),
    getNeedScores(),
    getInterests(),
    unreadCount(),
  ]);

  const name = profile.displayName || "there";

  return (
    <div className="dashboard-page">
      {unread > 0 && (
        <Link href="/dashboard/inbox" className="inbox-banner">
          You have {unread} unread message{unread > 1 ? "s" : ""}
        </Link>
      )}

      <div className="welcome-header">
        <h1>Hey {name}, how are things?</h1>
        <p>
          Keep this up to date and we&rsquo;ll connect you with people nearby
          who can help &mdash; and people you can help too.
        </p>
      </div>

      <ProfileBasics initial={profile} />

      <InterestEditor initial={interests} />

      <NeedSliders initial={scores} />
    </div>
  );
}
