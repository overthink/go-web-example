import Head from "next/head";
import { ReactNode } from "react";

interface Props {
  children: ReactNode;
}

const Layout = ({ children }: Props) => {
  return (
    <>
      <Head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>Tasks Admin</title>
      </Head>
      <nav className="navbar">
        <div className="navbar-menu">
          <div className="navbar-start">
            <a className="navbar-item">Home</a>
            <a className="navbar-item">Elsewhere</a>
            <a className="navbar-item">Whatev</a>
          </div>
        </div>
      </nav>
      <section className="section">
        <div className="container">{children}</div>
      </section>
    </>
  );
};

export default Layout;
