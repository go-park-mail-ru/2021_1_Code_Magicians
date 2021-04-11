--
-- PostgreSQL database dump
--

-- Dumped from database version 12.5 (Ubuntu 12.5-0ubuntu0.20.04.1)
-- Dumped by pg_dump version 12.5 (Ubuntu 12.5-0ubuntu0.20.04.1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: boards; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.boards (
    boardid integer NOT NULL,
    userid bigint NOT NULL,
    title character varying(100) NOT NULL,
    description text
);


ALTER TABLE public.boards OWNER TO postgres;

--
-- Name: TABLE boards; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.boards IS 'Boards that users have created';


--
-- Name: boards_boardid_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.boards_boardid_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.boards_boardid_seq OWNER TO postgres;

--
-- Name: boards_boardid_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.boards_boardid_seq OWNED BY public.boards.boardid;


--
-- Name: pins; Type: TABLE; Schema: public; Owner: postgres
--
-- DROP TABLE public.pins
CREATE TABLE public.pins (
    pinid integer NOT NULL,
    title character varying(100) NOT NULL,
    imagelink character varying(50) NOT NULL,
    description text
);


ALTER TABLE public.pins OWNER TO postgres;

--
-- Name: TABLE pins; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.pins IS 'Pins that users added';


--
-- Name: pins_pinid_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.pins_pinid_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.pins_pinid_seq OWNER TO postgres;

--
-- Name: pins_pinid_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.pins_pinid_seq OWNED BY public.pins.pinid;

----------------------------------------------------
CREATE TABLE public.pairs
(
    boardid integer NOT NULL,
    pinid   integer NOT NULL
);


ALTER TABLE public.pairs
    OWNER TO postgres;

--
-- Name: TABLE boards; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.pairs IS 'Pairs board-pin that users have created';
--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--
----------------------------------------------------------------
CREATE TABLE public.users (
    username character varying(45) NOT NULL,
    passwordhash character varying(40) NOT NULL,
    salt character(8) NOT NULL,
    email character varying(254) NOT NULL,
    last_name character varying(42),
    first_name character varying(42),
    avatar character varying(70) DEFAULT '/assets/img/default-avatar.jpg'::character varying NOT NULL,
    userid integer NOT NULL
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: TABLE users; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.users IS 'Table with all user data';


--
-- Name: users_userid_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.users_userid_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_userid_seq OWNER TO postgres;

--
-- Name: users_userid_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.users_userid_seq OWNED BY public.users.userid;


--
-- Name: boards boardid; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.boards ALTER COLUMN boardid SET DEFAULT nextval('public.boards_boardid_seq'::regclass);


--
-- Name: pins pinid; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pins ALTER COLUMN pinid SET DEFAULT nextval('public.pins_pinid_seq'::regclass);


--
-- Name: users userid; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN userid SET DEFAULT nextval('public.users_userid_seq'::regclass);


--
-- Data for Name: boards; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.boards (boardid, userid, title, description) FROM stdin;
\.


--
-- Data for Name: pins; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.pins (pinid, title, imagelink, description) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (username, passwordhash, salt, email, last_name, first_name, avatar, userid) FROM stdin;
tester	password123	        	email@mail.l			/assets/img/default-avatar.jpg	39
cupidatat_Duis_exercitation	officia consectetur irure	        	dgrDmB2TdK@VX.tvs			avatars/vmoBwFUKQyNirHZ_0b5OoarW6y4F1XpNAwfgRJrVZsODjdIbat4OIw==.jpg	40
\.


--
-- Name: boards_boardid_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.boards_boardid_seq', 1, false);


--
-- Name: pins_pinid_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.pins_pinid_seq', 1, false);


--
-- Name: users_userid_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_userid_seq', 40, true);


--
-- Name: boards boards_pk_oardid; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.boards
    ADD CONSTRAINT boards_pk_oardid PRIMARY KEY (boardid);


--
-- Name: pins pins_pk_pinid; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pins
    ADD CONSTRAINT pins_pk_pinid PRIMARY KEY (pinid);


--
-- Name: users users_pk_id; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pk_id PRIMARY KEY (userid);


--
-- Name: users users_un_email; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_un_email UNIQUE (email);


--
-- Name: users users_un_username; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_un_username UNIQUE (username);


--
-- Name: users_un_avatar; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX users_un_avatar ON public.users USING btree (avatar) WHERE ((avatar)::text <> '/assets/img/default-avatar.jpg'::text);


--
-- Name: boards boards_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.boards
    ADD CONSTRAINT boards_fk FOREIGN KEY (userid) REFERENCES public.users(userid) ON UPDATE CASCADE ON DELETE CASCADE;


ALTER TABLE ONLY public.pairs
    ADD CONSTRAINT pairs_fk FOREIGN KEY (boardid) REFERENCES public.boards(boardid)  ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: pins pins_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

-- ALTER TABLE ONLY public.pins
--     ADD CONSTRAINT pins_fk FOREIGN KEY (boardid) REFERENCES public.boards(boardid) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

