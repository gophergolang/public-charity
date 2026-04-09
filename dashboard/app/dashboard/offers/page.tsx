import { getOffers } from "@/lib/actions/offers";

export default async function OffersPage() {
  const offerList = await getOffers();

  return (
    <main>
      <h1>Your Offers</h1>
      {offerList.length === 0 ? (
        <p className="muted">You haven't added any offers yet.</p>
      ) : (
        <div className="offer-list">
          {offerList.map((o) => (
            <div key={o.id} className="card">
              <span className="tag">{o.category}</span>
              <p>{o.description}</p>
            </div>
          ))}
        </div>
      )}
    </main>
  );
}
