CREATE TABLE lang(
    three char(3) PRIMARY KEY,
    two char(2) UNIQUE,
    eng_desc varchar(20) NOT NULL
);

CREATE TABLE wikidata(
    id varchar(10) PRIMARY KEY,
    commons varchar(60) UNIQUE,
    eng_lang char(2) REFERENCES lang(two),
    eng_slug varchar(60),
    CHECK (eng_lang IS NOT NULL AND eng_slug IS NOT NULL OR eng_lang IS NULL AND eng_slug IS NULL)
);

CREATE TABLE author(
    slug varchar(20) PRIMARY KEY CONSTRAINT lowercase_or_minus CHECK (slug ~ '[a-z-]'),
    page bool NOT NULL,
    lang char(3) REFERENCES lang(three) NOT NULL,
    birth integer,
    death integer,
    wikidata varchar(10) UNIQUE REFERENCES wikidata(id),
    onlinebooks varchar(80) UNIQUE
);

CREATE TABLE name(
    author varchar(20) REFERENCES author(slug),
    site_lang char(3) REFERENCES lang(three),
    first_part varchar(40),
    main_part varchar(20) NOT NULL,
    last_part varchar(20),
    PRIMARY KEY(author, site_lang)
);

CREATE TABLE work(
    slug varchar(40) PRIMARY KEY CONSTRAINT lowercase_or_minus CHECK (slug ~ '[a-z-]'),
    page bool NOT NULL,
    lang char(3) REFERENCES lang(three) NOT NULL,
    translation varchar(40) REFERENCES work(slug),
    wikidata varchar(10) UNIQUE REFERENCES wikidata(id)
);

CREATE TABLE title(
    work_slug varchar(40) REFERENCES work(slug),
    site_lang char(3) REFERENCES lang(three),
    first_part varchar(60),
    main_part varchar(70) NOT NULL,
    last_part varchar(560),
    PRIMARY KEY(work_slug, site_lang)
);

CREATE TABLE attribution(
    author_slug varchar(20) REFERENCES author(slug),
    work_slug varchar(40) REFERENCES work(slug),
    dubious bool NOT NULL,
    PRIMARY KEY(author_slug, work_slug)
);

CREATE TABLE edition(
    work_slug varchar(40) REFERENCES work(slug),
    year integer,
    important bool NOT NULL,
    description varchar(240) NOT NULL,
    PRIMARY KEY(work_slug, year)
);

CREATE TABLE quality(
    quality varchar(20) PRIMARY KEY
);

CREATE TABLE website(
    sitename varchar(20) PRIMARY KEY,
    domain varchar(120) NOT NULL,
    url varchar(120),
    label varchar(60) NOT NULL
);

CREATE TABLE source(
    sitename varchar(20) REFERENCES website(sitename),
    url varchar(120),
    quality varchar(20) REFERENCES quality(quality) NOT NULL,
    download bool NOT NULL,
    description varchar(240),
    PRIMARY KEY(sitename, url)
);

CREATE TABLE link_content(
    work_slug varchar(40),
    year integer,
    sitename varchar(20),
    url varchar(120),
    FOREIGN KEY (work_slug, year) REFERENCES edition (work_slug, year),
    FOREIGN KEY (sitename, url) REFERENCES source (sitename, url),
    PRIMARY KEY(work_slug, year, sitename, url)
);

