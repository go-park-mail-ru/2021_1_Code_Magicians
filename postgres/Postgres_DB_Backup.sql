--
-- PostgreSQL database dump
--

-- Dumped from database version 12.7 (Debian 12.7-1.pgdg100+1)
-- Dumped by pg_dump version 12.7 (Debian 12.7-1.pgdg100+1)

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

ALTER TABLE ONLY public.reports DROP CONSTRAINT reports_fk_1;
ALTER TABLE ONLY public.reports DROP CONSTRAINT reports_fk;
ALTER TABLE ONLY public.pairs DROP CONSTRAINT pairs_fk;
ALTER TABLE ONLY public.followers DROP CONSTRAINT followers_users_follower;
ALTER TABLE ONLY public.followers DROP CONSTRAINT followers_users_followed;
ALTER TABLE ONLY public.comments DROP CONSTRAINT comments_user_fk;
ALTER TABLE ONLY public.comments DROP CONSTRAINT comments_pin_fk;
ALTER TABLE ONLY public.boards DROP CONSTRAINT boards_fk;
DROP INDEX public.users_vk_id_idx;
DROP INDEX public.users_un_avatar;
ALTER TABLE ONLY public.users DROP CONSTRAINT users_un_username;
ALTER TABLE ONLY public.users DROP CONSTRAINT users_un_email;
ALTER TABLE ONLY public.boards DROP CONSTRAINT users_un_boards;
ALTER TABLE ONLY public.users DROP CONSTRAINT users_pk_id;
ALTER TABLE ONLY public.pins DROP CONSTRAINT pins_pk_pinid;
ALTER TABLE ONLY public.reports DROP CONSTRAINT one_pin_per_sender;
ALTER TABLE ONLY public.followers DROP CONSTRAINT followers_pk;
ALTER TABLE ONLY public.comments DROP CONSTRAINT comments_pk_id;
ALTER TABLE ONLY public.boards DROP CONSTRAINT boards_pk_oardid;
ALTER TABLE public.users ALTER COLUMN userid DROP DEFAULT;
ALTER TABLE public.reports ALTER COLUMN reportid DROP DEFAULT;
ALTER TABLE public.pins ALTER COLUMN pinid DROP DEFAULT;
ALTER TABLE public.comments ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.boards ALTER COLUMN boardid DROP DEFAULT;
DROP SEQUENCE public.users_userid_seq;
DROP TABLE public.users;
DROP SEQUENCE public.reports_reportid_seq;
DROP TABLE public.reports;
DROP SEQUENCE public.pins_pinid_seq;
DROP TABLE public.pins;
DROP TABLE public.pairs;
DROP TABLE public.followers;
DROP SEQUENCE public.comments_id_seq;
DROP TABLE public.comments;
DROP SEQUENCE public.boards_boardid_seq;
DROP TABLE public.boards;
SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: boards; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.boards (
                               boardid integer NOT NULL,
                               userid bigint NOT NULL,
                               title character varying(100) NOT NULL,
                               description text,
                               imagelink character varying(70) DEFAULT 'assets/img/default-board-avatar.jpg'::character varying NOT NULL,
                               imageheight integer DEFAULT 480 NOT NULL,
                               imagewidth integer DEFAULT 1200 NOT NULL,
                               imageavgcolor character(6) DEFAULT '5a5a5a'::bpchar NOT NULL
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
                                 id integer NOT NULL,
                                 text text NOT NULL
);


ALTER TABLE public.comments OWNER TO postgres;

--
-- Name: comments_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.comments_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.comments_id_seq OWNER TO postgres;

--
-- Name: comments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.comments_id_seq OWNED BY public.comments.id;


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
                             imagelink character varying(70) NOT NULL,
                             description text,
                             userid integer,
                             imageheight integer DEFAULT 0 NOT NULL,
                             imagewidth integer DEFAULT 0 NOT NULL,
                             imageavgcolor character(6) DEFAULT 'FFFFFF'::bpchar NOT NULL,
                             creationdate timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
                             reports_count integer DEFAULT 0 NOT NULL
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
-- Name: reports; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.reports (
                                reportid integer NOT NULL,
                                pinid integer NOT NULL,
                                senderid integer NOT NULL,
                                description text NOT NULL
);


ALTER TABLE public.reports OWNER TO postgres;

--
-- Name: reports_reportid_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.reports_reportid_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.reports_reportid_seq OWNER TO postgres;

--
-- Name: reports_reportid_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.reports_reportid_seq OWNED BY public.reports.reportid;


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
                              followed_by integer DEFAULT 0 NOT NULL,
                              pins_count integer DEFAULT 0 NOT NULL,
                              boards_count integer DEFAULT 0 NOT NULL,
                              vk_id integer DEFAULT 0 NOT NULL
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
-- Name: comments id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comments ALTER COLUMN id SET DEFAULT nextval('public.comments_id_seq'::regclass);


--
-- Name: pins pinid; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pins ALTER COLUMN pinid SET DEFAULT nextval('public.pins_pinid_seq'::regclass);


--
-- Name: reports reportid; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reports ALTER COLUMN reportid SET DEFAULT nextval('public.reports_reportid_seq'::regclass);


--
-- Name: users userid; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN userid SET DEFAULT nextval('public.users_userid_seq'::regclass);


--
-- Data for Name: boards; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.boards (boardid, userid, title, description, imagelink, imageheight, imagewidth, imageavgcolor) FROM stdin;
\.


--
-- Data for Name: comments; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.comments (userid, pinid, id, text) FROM stdin;
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

COPY public.pins (pinid, title, imagelink, description, userid, imageheight, imagewidth, imageavgcolor, creationdate, reports_count) FROM stdin;
\.


--
-- Data for Name: reports; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.reports (reportid, pinid, senderid, description) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (username, passwordhash, salt, email, last_name, first_name, avatar, userid, following, followed_by, pins_count, boards_count, vk_id) FROM stdin;
\.


--
-- Name: boards_boardid_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.boards_boardid_seq', 182, true);


--
-- Name: comments_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.comments_id_seq', 56, true);


--
-- Name: pins_pinid_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.pins_pinid_seq', 80, true);


--
-- Name: reports_reportid_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.reports_reportid_seq', 1, false);


--
-- Name: users_userid_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_userid_seq', 97, true);


--
-- Name: boards boards_pk_oardid; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.boards
    ADD CONSTRAINT boards_pk_oardid PRIMARY KEY (boardid);


--
-- Name: comments comments_pk_id; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pk_id PRIMARY KEY (id);


--
-- Name: followers followers_pk; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.followers
    ADD CONSTRAINT followers_pk PRIMARY KEY (followerid, followedid);


--
-- Name: reports one_pin_per_sender; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT one_pin_per_sender UNIQUE (pinid, senderid);


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
-- Name: boards users_un_boards; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.boards
    ADD CONSTRAINT users_un_boards UNIQUE (userid, title);


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
-- Name: users_vk_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX users_vk_id_idx ON public.users USING btree (vk_id)
WHERE NOT vk_id = 0;


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
-- Name: reports reports_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_fk FOREIGN KEY (pinid) REFERENCES public.pins(pinid) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: reports reports_fk_1; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_fk_1 FOREIGN KEY (senderid) REFERENCES public.users(userid) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--
