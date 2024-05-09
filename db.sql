create table university
(
  id   bigint auto_increment
    primary key,
  name varchar(255) collate utf8mb4_general_ci not null
)
  collate = utf8mb4_bin;

create table user
(
  username varchar(20)  not null
    primary key,
  password varchar(255) not null,
  role     char         not null,
  keyword  varchar(255) not null
)
  collate = utf8mb4_bin;

create table customer
(
  id       bigint auto_increment
    primary key,
  fname    varchar(30) not null,
  lname    varchar(30) not null,
  state    varchar(30) not null,
  city     varchar(2)  not null,
  zip      varchar(5)  not null,
  address  varchar(50) not null,
  username varchar(20) null,
  constraint customer_user_username_fk
    foreign key (username) references user (username)
      on update cascade on delete cascade
)
  collate = utf8mb4_bin;

create table account
(
  number    bigint      not null
    primary key,
  fname     varchar(30) not null,
  lname     varchar(30) not null,
  state     varchar(30) not null,
  city      varchar(2)  not null,
  zip       varchar(5)  not null,
  open_date datetime    not null,
  id        bigint      not null,
  address   varchar(50) not null,
  type      char        not null,
  constraint account_customers_id_fk
    foreign key (id) references customer (id)
      on update cascade on delete cascade,
  constraint type_check
    check (`type` in (_utf8mb4\'L\',_utf8mb4\'C\',_utf8mb4\'S\'))
)
collate=utf8mb4_bin;

create table checking
(
    number  bigint                               not null
        primary key,
    charge  decimal(7, 2)                        not null,
    balance decimal(15, 2) unsigned default 0.00 not null,
    constraint checking_account_number_fk
        foreign key (number) references account (number)
            on update cascade on delete cascade
)
    collate = utf8mb4_bin;

create table loan
(
    number  bigint         not null
        primary key,
    rate    decimal(5, 2)  not null,
    amount  decimal(10, 2) not null,
    months  int            not null,
    payment decimal(10, 2) not null,
    type    varchar(8)     not null,
    constraint loan_account_number_fk
        foreign key (number) references account (number)
            on update cascade on delete cascade,
    constraint check_account_type
        check (`type` in (_utf8mb4\'STUDENT\',_utf8mb4\'HOME\',_utf8mb4\'PERSONAL\'))
)
collate=utf8mb4_bin;

create table home_loan
(
    number                    bigint         not null
        primary key,
    house_built_year          varchar(4)     not null,
    insurance_acc_no          bigint         not null,
    insurance_company_name    varchar(50)    not null,
    insurance_company_state   varchar(30)    not null,
    insurance_company_city    varchar(2)     not null,
    insurance_company_address varchar(50)    not null,
    insurance_company_zip     varchar(5)     not null,
    insurance_company_premium decimal(10, 2) not null,
    constraint home_load_pk_2
        unique (insurance_acc_no),
    constraint home_loan_loan_number_fk
        foreign key (number) references loan (number)
            on update cascade on delete cascade
)
    collate = utf8mb4_bin;

create table loan_payment
(
    id             bigint auto_increment
        primary key,
    number         bigint                  not null,
    payment_date   date                    not null,
    payment_amount decimal(10, 2) unsigned not null,
    constraint fk_loan_number
        foreign key (number) references loan (number)
            on update cascade on delete cascade
)
    collate = utf8mb4_bin;

create table savings
(
    number  bigint                               not null
        primary key,
    rate    decimal(5, 2)                        not null,
    balance decimal(15, 2) unsigned default 0.00 not null,
    constraint savings_account_number_fk
        foreign key (number) references account (number)
            on update cascade on delete cascade
)
    collate = utf8mb4_bin;

create table student_loan
(
    number           bigint      not null
        primary key,
    university_id    bigint      not null,
    student_id       varchar(30) not null,
    education_type   char        not null,
    expect_grad_date date        not null,
    constraint student_id
        unique (student_id),
    constraint student_loan_loan_number_fk
        foreign key (number) references loan (number)
            on update cascade on delete cascade,
    constraint student_loan_university_id_fk
        foreign key (university_id) references university (id)
            on update cascade on delete cascade
)
    collate = utf8mb4_bin;

create table transfer_history
(
    id                  bigint auto_increment
        primary key,
    from_account_number bigint         not null,
    to_account_number   bigint         not null,
    account_type        char           not null,
    amount              decimal(15, 2) not null,
    transfer_date       datetime       not null,
    constraint fk_transfer_history_from_account_number
        foreign key (from_account_number) references account (number)
            on update cascade on delete cascade,
    constraint fk_transfer_history_to_account_number
        foreign key (to_account_number) references account (number)
            on update cascade on delete cascade
)
    collate = utf8mb4_bin;

