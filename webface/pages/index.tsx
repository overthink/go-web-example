import type { NextPage } from "next";
import Head from "next/head";

const Home: NextPage = () => {
  return (
    <>
      <Head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>Example webapp title</title>
      </Head>
      <nav className="navbar">
        <div className="navbar-menu">
          <div className="navbar-start">
            <a className="navbar-item">Home</a>
          </div>
          <div className="navbar-end">
            <a className="navbar-item">
              <button className="button is-primary">Sign in</button>
            </a>
          </div>
        </div>
      </nav>
      <section className="section">
        <div className="container">
          <div className="columns">
            <div className="column">Foo!</div>
            <div className="column">
              <button className="button is-primary">Push me</button>
            </div>
            <div className="column">Baz!</div>
          </div>
        </div>
      </section>
      <footer className="footer">
        <div className="content has-text-centered">
          Here is some footer text.
        </div>
      </footer>
    </>
  );
};

export default Home;
