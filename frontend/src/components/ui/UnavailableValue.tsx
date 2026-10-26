export const unavailableLabel='Data belum tersedia';
export function UnavailableValue({reason}:{reason:string}){return <span className="unavailable-value"><span>{unavailableLabel}</span><small>{reason}</small></span>}
