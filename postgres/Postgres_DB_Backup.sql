--
-- PostgreSQL database dump
--

-- Dumped from database version 13.2 (Debian 13.2-1.pgdg100+1)
-- Dumped by pg_dump version 13.2 (Debian 13.2-1.pgdg100+1)

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
    userid integer
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
-- Name: comments id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.comments ALTER COLUMN id SET DEFAULT nextval('public.comments_id_seq'::regclass);


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
130	80	bigblackboard	Amazingestest-board
131	81	Saved pins	Fast save
133	80	testboard	Amazingestest-board
134	81	Board2	Amazingestest-board
136	81	gdfdfgdf	Amazingestest-board
137	81	kkkkkkkkkkkk	Amazingestest-board
138	84	Saved pins	Fast save
139	80	qwe	Amazingestest-board
140	85	Saved pins	Fast save
141	86	Saved pins	Fast save
142	87	Saved pins	Fast save
143	88	Saved pins	Fast save
144	89	Saved pins	Fast save
145	90	Saved pins	Fast save
146	91	Saved pins	Fast save
147	92	Saved pins	Fast save
148	93	Saved pins	Fast save
149	93	Basddd	Amazingestest-board
150	94	Saved pins	Fast save
151	94	е39	Amazingestest-board
154	80	Saved pins	Amazingestest-board
156	95	Saved pins	Fast save
159	80	pinter-best-boy	Amazingestest-board
162	96	Saved pins	Fast save
172	90	culpa et occaecat ullamco tempor	eli
176	80	testing	
177	80	testing again	
178	80	readreadread read read read	
181	80	AlexBoardShow	
182	97	Saved pins	Fast save
\.


--
-- Data for Name: comments; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.comments (userid, pinid, id, text) FROM stdin;
81	30	17	Loll
81	30	18	Lol, nice new comment!\n
81	30	19	afad
81	30	20	afad
80	37	21	qwe
80	37	22	
80	37	23	lol
86	46	24	is it my comment?
80	40	25	Thinking
80	48	26	I could do a cup of coffee atm
80	48	27	Likewise
80	48	28	me three
90	56	29	Follow me now!!!!!!!!!!!!!\n
81	40	30	Loool, nice dog!\n
91	44	31	wow
87	63	33	Вау
87	63	34	ааа
80	62	35	I agree
91	62	36	((((::\n\n
91	62	37	
91	62	38	
91	62	39	
91	62	40	
91	62	41	
90	64	42	bmw for losers
95	66	43	test
80	66	44	nice cat
95	66	45	test1
95	66	46	Test2 dfsfdsdgsdgsdgsd
95	66	47	
95	66	48	
95	66	49	   
95	53	50	ddsvsdvsd
96	48	51	asdfsfsdfsd
96	66	52	
81	39	53	lallala
81	39	54	loh
81	30	55	comm
87	38	56	aaa
\.


--
-- Data for Name: followers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.followers (followerid, followedid) FROM stdin;
80	86
81	80
84	80
92	80
84	91
80	91
81	91
80	95
95	80
95	90
96	87
84	81
90	80
\.


--
-- Data for Name: pairs; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.pairs (boardid, pinid) FROM stdin;
131	30
131	31
131	32
131	33
131	34
131	35
131	36
141	46
145	56
147	61
146	62
142	63
150	64
145	65
156	66
145	69
145	70
145	71
182	72
182	73
182	74
182	75
182	76
182	77
182	78
182	79
182	80
\.


--
-- Data for Name: pins; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.pins (pinid, title, imagelink, description, userid) FROM stdin;
38	Kitty cat	pins/yDWzBLdXVuJm6IFL3AVbRXh4Q1MHA4irDpikg3qTjg-DjFfYY_NUGQ==.jpg	Awesomest cat	80
39	Corki	pins/xwV2LgOEHLz8CwfGPNZ8870q1o5hlkzdtd1E9CZGskroxy7xUq0fog==.jpg	Pupper a day keeps your fears away	80
40	Greatest thinker	pins/NM3KwICy7KMJlOw0GA0gJPZKmIRIiYk4H3y97oocjYfAFBHxjlcmrg==.jpg	My best thinking face	80
41	Wisdom	pins/pwL-FKUjnTQiB7KgK2ah-lztSk5sxC9dfo4usDQ6YMfM6hBkrCFqpw==.jpg	don't drink and text	80
42	Truth	pins/Gam79ZLsv1_qXFYZYnEstaTTMaE4k8WJbE-YlmB1qVqP_Ee-5paXiA==.jpg	:(	80
44	Quacken	pins/wRC4xQbth35DKBzczbjciqx0MCLjPASbxbFitfks_nZWOVdgNNdvoQ==.jpg	Test	80
45	Fort	pins/9vBT0B2gBwAuCxl5eA6JtU_Ncv6RiZImtG8S9yjMzCElVrcYVunOag==.jpg	Hold the fort	80
46	try to set name	pins/QUpfkXRVDeCZTwnViqFvDXesNlm_bU2bphstNKX6QvRhAKekmWC9dA==.jpg	desc	86
48	cup	pins/7K-qQCLFC4mb4VYRAIdlA-RHwUW9fIrvL6rmaPy1rOZ15V0jbmNUmg==.jpg	amazing coffee	80
49	sleep tight <3	pins/7CYLBWd6ASHWAcr9gDjB46Bxl4WLvW03yDMhFybcRr74NnTO3SSmaQ==.jpg	for my homies	80
50	too fast	pins/-Rm0vXDOHJ-HEm-e4btnUxDv9VBfcbRA4KGG32ebSuWP9jJAoWby6Q==.jpg	fastest	80
51	say hello	pins/3Fso5nMtlLB1gAZkG3lw--8LLtGD5s5TFjZr23VqrHwcrP65JlxRlQ==.jpg	Say hello to my little friend	80
52	relationship goals	pins/4dTdzpres7QqpWW4U1poAYmqGHB51hQ5M0-xPTun3Qv9X-hCPeRmOw==.jpg	if you aint this serious, I aint playin	80
53	the count	pins/zREDw5tItpyhWkbJGBnO6GFUSr6Yty5zol9VV3qmQ_jISh6hxm9wfw==.jpg	Immune to the counting vampire :)	80
54	lift	pins/-NoPtJEX8VRVpvE4TeFIDbzMd6dXhIGmxqaHgs8mecdaMcgCBwF6fA==.jpg	re: haha funny	80
55	surprise	pins/cnZ4DjefbpsPMkhocFDcp46_VSimNP9W74-QbBAnyIQwxDiYtELY-Q==.jpg	as mad as my future ex wife	80
56	My first project	pins/mVGo8iMVxSoW5CZp6MMgPVo4nZnty-Zlt53A-FtnzbhQ1433L0_Avw==.jpg	On C++	90
57	yes	pins/fG9yy8ZfqhyPi4klw5dX7Iy95X3bBlyQZ9YUC8z3soTq-6gAbCJzLA==.jpg	yes, I am just as cute :)	80
58	true fear	pins/KmGojBXh8RYChfjjUokvm--IHgsr8ce5LH7e5yrtaqaLLqyRNFUJiQ==.jpg	Don't be afraid!	80
71		pins/V23F4jD9f6w0yTRcviAGuZMclMFstjG0NxKA2wOMv_d3q-9Nm44wlg==.jpg		90
30	Pin123	pins/Uy3al5hS_GR117en5dreuCp61OEvv9XKzCuo6XWg0FtHOp4-fi6o6w==.jpg	podfsfsf	81
31	ddd	pins/ozQ9mb7wyHIlUIdXhzd8uy2XC1a_iy3J9gXA6_Q7X7Nkhx_S3GgbOQ==.jpg	fsad	81
32	TTITTTLE	pins/sNjeSG2KeT3uigH9B2FPS92jv_6_SrP3gCRNbce7ph5LTHVqbDsHWA==.jpg	fadssfff	81
59	Ctrl-C Ctrl-V	pins/EaaF76AnhtO8n6N8CXVHJ25OKruYkOKA8VCu2aeSgMdUbFVj2F8W0w==.jpg	wish it was that easy	80
36	ooo	pins/Yk179wdfn1E6LW_kn6eXU2eMODxWlzMIDhM5_J1MDL5eAVhuJf60yQ==.jpg	mmjkj	81
37	qwe	pins/xlAjXAcBjQPHl8Lh8M1hKnODf6zfToPR1tRnWzL-6ZKAlfVaMU6pnw==.jpg	qwe	80
60	test	pins/4_GwkP8GSDBg3zaxgHRw4uke28HfaFDNduGIO1yJrc6GI48TOHIWsQ==.jpg	rl	80
62	TESTING AH	pins/DPGPAzz643ZCDjLWPnYSZKm__6fYlxmfSXWwCcAxGqo4KxSqErWq9w==.jpg	BEST PICTURE IN PINTER-BEST	91
63	Зеленый фон	pins/4ds8g6gP5qwGUk2vvwzOtXl5sJhl-XAycEFKb6QVbTCvySrf6-N5RQ==.jpg	Красивый фон, вай красота	87
64	мага	pins/oChMzR7KS8Up-6uCb41tlAPGEKL-BUT27DQS6cJixUOB25TP0pPg9w==.jpg	бавариан мотор вэрке	94
65	Курсач	pins/s1rRJwZm8-cNi1L13o2gTVrwSgBGJoFoaCR_28Mv89vUYCFRTa1Kdw==.jpg	python	90
66	Cat	pins/KRrhCB4gq7bqPHhI7W7qfjMnRUKvAkk2b79Qou0z1mqafkTUWjiEdw==.jpg	Cat-mem	95
67	kj	pins/yv6dhwBSFsfaHleilefwApHsUSWMqXtoC634xHuOIGnIlW2ddpxwxw==.jpg	python	90
69	incididunt	pins/ybHJp2H9WQWsjVz6V7llTrWkBgZO-sQnudA2M8WnK9ktsw7u80AXoA==.jpg	ipsum exercitation	90
72	232sss	pins/U-f0NhVD64z9FU8AlygkZ-_EFND93A5r5Db6aarFbv3M-qnlZJSutA==.jpg	asdf	97
74	APACHE	pins/vm07JCBDjeTWOFVVieUClHhUXD3j2oB42J-A9QIFb0-lHMW5bIClrw==.jpg	lalala	97
78	AWS	pins/Y77vWC2OkP2VdQ4LU7_o-9PxUdD-Yq_bbj8mHXTiw6oCjn6efa5-xQ==.jpg	normaaaaalno	97
80	prig	pins/CbFZbHakonwt2bnCHBV8O_4ckDMXPhc65Uy432h05m18K-PyeslH_g==.jpg	skok	97
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (username, passwordhash, salt, email, last_name, first_name, avatar, userid, following, followed_by) FROM stdin;
asdfasdf	asasdfasdf	        	asdfasdf@fasdkfjasjk.ru	\N	\N	assets/img/default-avatar.jpg	92	1	0
IIIIIffff	hokmas-8myczu-jEtkaq	        	asdasdasd@gmffail.cfffom		wowfff	assets/img/default-avatar.jpg	91	0	3
Name	Qwerty1234	        	vdsf@lkjd.com	\N	\N	assets/img/default-avatar.jpg	93	0	0
Akhmat	10188eпщк	        	egor10188@mail.ru		username	assets/img/default-avatar.jpg	94	0	0
admin	0987654321	        	admin@admin.ru		ADmin	avatars/tKbjq0ZmhkHenQ7R_Vgpi52nFmVf61YRA23hYeieYxx8r9ZpWeX2Sg==.jpg	87	0	1
tester	password123	        	testes@tsd.omm	\N	\N	assets/img/default-avatar.jpg	84	3	0
Naum	123456789kek	        	vvbh@mail.ru	\N	\N	assets/img/default-avatar.jpg	85	0	0
DmitriyR	oop3jaug8U	        	romashov@gmail.com		Dmitriy	avatars/X9FpwMn8zaNLxyeCdXuEMaLG1HSI-q3Rd9-hkq9ke_ZxSJ4SghUuKg==.jpg	95	2	1
michael	1234512345	        	mail@mail.ru	\N	\N	assets/img/default-avatar.jpg	86	0	1
randomUser	qweqweqwe	        	qwe@qwe.qwe	\N	\N	assets/img/default-avatar.jpg	88	0	0
random	qweqweqwe	        	qwe@qwe.qe	\N	\N	assets/img/default-avatar.jpg	89	0	0
nikitalol123	paofasdfpj	        	erere@rdrs.com			avatars/aeFUFlMhpi1ELmK7NpqAIDcfQ30hnEUjMje-YTffliDLqJuh9Vfp3A==.jpg	96	1	0
nikitaadadada	password123	        	password@fgfas.fasd			avatars/iOJW8k-TyF5eKNVboUqBjPf-_IjglOwhMvTd22dd1NGygwGAZbk_1w==.jpg	81	2	1
nikita	password123	        	ermilov1999@gmail.com	\N	\N	assets/img/default-avatar.jpg	97	0	0
Naum1	1234567890	        	yashuvaevni@student.bmstu.ru			avatars/Emhrn9GWYR5P7uoAC1xy4cXEoJ4enZYFFoEm48wZ5xhFq_bf4LNtyA==.jpg	90	1	1
vk	qweqweqwe	        	vk@vk.vk	\N	\N	assets/img/default-avatar.jpg	80	3	5
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

