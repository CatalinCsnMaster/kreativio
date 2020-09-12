create table public.images (
    id serial not null primary key,
    created_at timestamp with time zone default now(),
    link_original text,
    link_resized text
);
