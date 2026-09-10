import type { ReactNode } from "react";

export function PageHeader({ title, detail, breadcrumb, action }: { title: string; detail: string; breadcrumb?: ReactNode; action?: ReactNode }) {
  return <header className="page-header">{breadcrumb ? <nav className="breadcrumb" aria-label="Breadcrumb">{breadcrumb}</nav> : null}<div className="page-header__row"><div><h1>{title}</h1><p>{detail}</p></div>{action}</div></header>;
}
