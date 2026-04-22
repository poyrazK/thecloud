
'use client';

import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { Cloud, LayoutGrid, Server, HardDrive, Network, Settings, Activity, Menu, X } from 'lucide-react';
import styles from './Sidebar.module.css';

const MENU_ITEMS = [
  { name: 'Dashboard', icon: LayoutGrid, href: '/dashboard' },
  { name: 'Compute', icon: Server, href: '/compute' },
  { name: 'Storage', icon: HardDrive, href: '/storage' },
  { name: 'Network', icon: Network, href: '/network' },
  { name: 'Activity', icon: Activity, href: '/activity' },
  { name: 'Settings', icon: Settings, href: '/settings' },
];

export const Sidebar: React.FC = () => {
  const pathname = usePathname();
  const router = useRouter();
  const [open, setOpen] = useState(false);

  useEffect(() => {
    MENU_ITEMS.forEach((item) => {
      router.prefetch(item.href);
    });
  }, [router]);

  const handleNavigate = (href: string) => (event: React.MouseEvent<HTMLAnchorElement>) => {
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey || event.button !== 0) {
      setOpen(false);
      return;
    }

    setOpen(false);

    if (pathname === href) {
      event.preventDefault();
      return;
    }

    event.preventDefault();
    router.push(href);
  };

  return (
    <>
      <button
        className={styles.mobileToggle}
        type="button"
        onClick={() => setOpen((current) => !current)}
        aria-label="Toggle navigation"
      >
        {open ? <X size={18} /> : <Menu size={18} />}
        <span>Console</span>
      </button>

      {open ? (
        <button
          type="button"
          className={`${styles.backdrop} ${styles.backdropVisible}`}
          onClick={() => setOpen(false)}
          aria-label="Close navigation"
        />
      ) : null}

      <aside className={`${styles.sidebar} material-sidebar ${open ? styles.mobileOpen : ''}`}>
        <div className={styles.logo}>
          <div className={styles.logoIcon}>
            <Cloud size={16} />
          </div>
          <div>
            <div className={styles.logoText}>The Cloud</div>
            <div className={styles.logoSub}>Control Console</div>
          </div>
        </div>

        <nav className={styles.nav}>
          {MENU_ITEMS.map((item) => {
            const Icon = item.icon;
            const isActive = pathname === item.href;

            return (
              <Link
                key={item.name}
                href={item.href}
                prefetch
                className={`${styles.navItem} ${isActive ? styles.active : ''}`}
                onClick={handleNavigate(item.href)}
              >
                <Icon size={16} strokeWidth={2.1} style={{ opacity: isActive ? 1 : 0.78 }} />
                <span>{item.name}</span>
              </Link>
            );
          })}
        </nav>

        <div className={styles.footer}>
          <div className={styles.status}>
            <div className={styles.statusDot} />
            <span>Region: us-east-1</span>
          </div>
          <p className={styles.footerText}>Open-source cloud runtime for compute, storage, and networking.</p>
        </div>
      </aside>
    </>
  );
};
