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
-- Name: comments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.comments (
    userid integer NOT NULL,
    pinid integer NOT NULL,
    text text NOT NULL
);


ALTER TABLE public.comments OWNER TO postgres;

--
-- Name: followers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.followers (
    followerid integer NOT NULL,
    followedid integer NOT NULL
);


ALTER TABLE public.followers OWNER TO postgres;

--
-- Name: COLUMN followers.followerid; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.followers.followerid IS 'User who follows';


--
-- Name: COLUMN followers.followedid; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.followers.followedid IS 'User who is followed';


--
-- Name: pairs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.pairs (
    boardid integer NOT NULL,
    pinid integer NOT NULL
);


ALTER TABLE public.pairs OWNER TO postgres;

--
-- Name: TABLE pairs; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.pairs IS 'Pairs board-pin that users have created';


--
-- Name: pins; Type: TABLE; Schema: public; Owner: postgres
--

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


--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    username character varying(45) NOT NULL,
    passwordhash character varying(40) NOT NULL,
    salt character(8) NOT NULL,
    email character varying(254) NOT NULL,
    last_name character varying(42),
    first_name character varying(42),
    avatar character varying(70) DEFAULT 'assets/img/default-avatar.jpg'::character varying NOT NULL,
    userid integer NOT NULL,
    following integer DEFAULT 0 NOT NULL,
    followed_by integer DEFAULT 0 NOT NULL
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
10	71	Saved pins	
11	72	Saved pins	
\.


--
-- Data for Name: comments; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.comments (userid, pinid, text) FROM stdin;
\.


--
-- Data for Name: followers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.followers (followerid, followedid) FROM stdin;
\.


--
-- Data for Name: pairs; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.pairs (boardid, pinid) FROM stdin;
\.


--
-- Data for Name: pins; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.pins (pinid, title, imagelink, description) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (username, passwordhash, salt, email, last_name, first_name, avatar, userid, following, followed_by) FROM stdin;
cupidatat_Duis_exercitation	officia consectetur irure	        	dgrDmB2TdK@VX.tvs	\N	\N	assets/img/default-avatar.jpg	72	0	0
example	w5t5rhdsgerthtyretsdr	        	other@mail.ru	\N	\N	assets/img/default-avatar.jpg	71	0	0
\.


--
-- Name: boards_boardid_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.boards_boardid_seq', 11, true);


--
-- Name: pins_pinid_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.pins_pinid_seq', 1, false);


--
-- Name: users_userid_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_userid_seq', 72, true);


--
-- Name: boards boards_pk_oardid; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.boards
    ADD CONSTRAINT boards_pk_oardid PRIMARY KEY (boardid);


--
-- Name: followers followers_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.followers
    ADD CONSTRAINT followers_pk PRIMARY KEY (followerid, followedid);


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

CREATE UNIQUE INDEX users_un_avatar ON public.users USING btree (avatar) WHERE ((avatar)::text <> 'assets/img/default-avatar.jpg'::text);


--
-- Name: boards boards_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.boards
    ADD CONSTRAINT boards_fk FOREIGN KEY (userid) REFERENCES public.users(userid) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: comments comments_pin_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pin_fk FOREIGN KEY (pinid) REFERENCES public.pins(pinid) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: comments comments_user_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_user_fk FOREIGN KEY (userid) REFERENCES public.users(userid) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: followers followers_users_followed; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.followers
    ADD CONSTRAINT followers_users_followed FOREIGN KEY (followedid) REFERENCES public.users(userid);


--
-- Name: followers followers_users_follower; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.followers
    ADD CONSTRAINT followers_users_follower FOREIGN KEY (followerid) REFERENCES public.users(userid);


--
-- Name: pairs pairs_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pairs
    ADD CONSTRAINT pairs_fk FOREIGN KEY (boardid) REFERENCES public.boards(boardid) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

