create table users (
    accountusername VARCHAR UNIQUE PRIMARY KEY NOT NULL,
    accountpassword VARCHAR NOT NULL
);

insert into users values ('nadine', 'nadine');
insert into users values ('nathan', 'nathan');
insert into users values ('terrell', 'terrell');
insert into users values ('garner', 'garner');