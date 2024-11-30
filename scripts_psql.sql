-- change user_schools role parents to user_parents

UPDATE user_parents up
SET school_uid = us.school_uid
FROM user_schools us
WHERE us.user_uid = up.child_uid;

DELETE FROM user_schools WHERE role_code = 'parent';


-- delete secondary lessons (subject.parent_uid is not null)
with ss as (
select l.uid as maxid from subjects s 
left join lessons l on l.subject_uid=s.uid
where s.parent_uid is not null and l.title is null
) 
delete from lessons using ss
where uid=maxid;

-- migration: delete same user_classrooms (have to run multiple times)
ALTER TABLE user_classrooms ADD COLUMN id SERIAL PRIMARY KEY;

with ss as (
select min(id) as minid ,max(id) as maxuid, user_uid, classroom_uid 
from user_classrooms
where type is null
group by user_uid, classroom_uid
having count(id) > 1)
delete  from user_classrooms us using ss
where id=maxuid ;

alter table user_classrooms drop column id;


-- delete sub classrooms without main
select count(*)--- delete 
from user_classrooms u
where type_key is not null and NOT EXISTS (SELECT 1 FROM user_classrooms t WHERE t.user_uid=u.user_uid and t.classroom_uid=u.classroom_uid and t.type is null);


-- add documents key, restore old cols

update users set documents=concat('[{"key":"birth_certificate","number":"',birth_cert_number,'","date":null},{"key":"order_number","number":"',apply_number,'","date":null},{"key":"passport","number":"',passport_number,'","date":null}]')::json
where documents::text='[]';


update users set apply_number=replace(apply_number,'"', ' ') where apply_number like '%"%';
update users set passport_number=replace(passport_number,'"', ' ') where passport_number like '%"%';
update users set birth_cert_number=replace(apply_number,'"', ' ') where birth_cert_number like '%"%';

update users set apply_number=replace(apply_number,'	', ' ') where apply_number like '%	%';
update users set passport_number=replace(passport_number,'	', ' ') where passport_number like '%	%';
update users set birth_cert_number=replace(birth_cert_number,'	', ' ') where birth_cert_number like '%	%';

update users set apply_number=replace(apply_number,'\', ' ') where apply_number like '%\%';
update users set passport_number=replace(passport_number,'\', ' ') where passport_number like '%\%';
update users set birth_cert_number=replace(birth_cert_number,'\', ' ') where birth_cert_number like '%\%';



-- reset usernames

update users set username=REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(LOWER(username), 'ý', 'y'), 'ö', 'o'), 'ä', 'a'), 'ň', 'n'), 'ü', 'u'), 'ç', 'c');


-- case upgrade classrooms

update classrooms set name=trim(name);

select name, CONCAT((substr(name, 1, length(name)-1)::integer-1)::text,RIGHT(name, 1)) 
	from classrooms 
	where length(name) BETWEEN 2 and 3 and RIGHT(name,1) ~ '^[0-9\.]+$' = false and school_uid='ca3220a8-8fc1-4bff-b512-ab2061c3b9ef' limit 100;

update classrooms set name=CONCAT((substr(name, 1, length(name)-1)::integer-1)::text,RIGHT(name, 1)) 
	where school_uid='ca3220a8-8fc1-4bff-b512-ab2061c3b9ef' and length(name) BETWEEN 2 and 3 and RIGHT(name,1) ~ '^[0-9\.]+$' = false;



-- migration: delete same parent child
select count(*) from user_parents up where up.parent_id =up.child_id;
delete from user_parents up where up.parent_id =up.child_id;

-- classroom to Turkmen
update classrooms
set name=CONCAT(replace(name,'B!^',''),'Ç')
where name like '%B!^%';

update classrooms
set name=CONCAT(replace(name,'D!^',''),'Ä')
where name like '%D!^%';


---

alter table user_classrooms drop CONSTRAINT user_classrooms_unique;

ALTER TABLE user_classrooms ADD CONSTRAINT "user_classrooms_unique" 
UNIQUE NULLS NOT DISTINCT  (user_uid, classroom_uid, type);

ALTER TABLE user_classrooms ADD CONSTRAINT "user_classrooms_unique" 
UNIQUE (user_uid, classroom_uid, type);

ALTER TABLE user_classrooms ADD CONSTRAINT "user_classrooms_unique_classroom"  UNIQUE (user_uid, classroom_uid, type, type_key);

-- migration: delete same user_schools (have to run multiple times)

ALTER TABLE user_schools ADD COLUMN id SERIAL PRIMARY KEY;

with ss as (
select min(id) as minid ,max(id) as maxuid, user_uid, school_uid, role_code  
from user_schools
group by user_uid, school_uid, role_code
having count(id) > 1) 
delete  from user_schools us using ss
where id=maxuid ;

alter table user_schools drop column id;

ALTER TABLE user_schools ADD CONSTRAINT "user_schools_unique" UNIQUE (user_uid, school_uid, role_code);



-- migration: set subjects.parent_id (set same subjects in class)
with ss as (
select min(id) as minid ,max(id) as maxid, name, classroom_id  
from subjects
group by name, classroom_id
having count(id) > 1
) 
UPDATE subjects
SET parent_id=ss.minid
from ss 
WHERE id=ss.maxid;

-- migration: delete same period_grades (have to run multiple times)
with ss as (
select min(id) as minid ,max(id) as maxid, student_id, subject_id, period_key  
from period_grades
group by student_id, subject_id, period_key 
having count(id) > 1 
) 
delete  from period_grades pg using ss
where id=maxid ;


-- migration: update period_ids (NOT TESTED)
with ss as (
select id, max(p.id) as period_id
from period_grades pg
right join subjects s on s.id=pg.subject_id
right join periods p on p.school_id=s.school_id
group by id
) 
UPDATE period_grades
SET period_id=ss.period_id
from ss
WHERE id=ss.id;

-- migration: update or insert given period grades
insert into period_grades (period_id, period_key, subject_id, student_id, lesson_count, absent_count, grade_count, grade_sum, updated_at, created_at)
	(select max(p.id), 3, max(sb.id) as subject_id, uc.user_id as student_id, count(distinct l.id) as lesson_count, count(a.id) as absent_count,
	count(g.id) as grade_count, 
	COALESCE(SUM(g.value::int)+(SUM(COALESCE(values[1]::int, 0))+SUM(COALESCE(values[2]::int, 0)))/2,0) as grade_sum, current_timestamp, current_timestamp
	 from subjects sb 
		left join classrooms c on c.id=sb.classroom_id left join user_classrooms uc on uc.classroom_id=c.id and uc.type_key is null
		left join lessons l on (l.subject_id=sb.id and l.period_key=3)
		left join absents a on (a.lesson_id=l.id and a.student_id=uc.user_id)
		left join grades g on (g.lesson_id=l.id and g.student_id=uc.user_id)
		right join users u on (u.id=uc.user_id)
		right join periods p on (p.school_id=sb.school_id)
		where sb.parent_id is null and uc.user_id is not null
		group by sb.id, uc.user_id )
	ON CONFLICT(student_id,subject_id,period_key) 
	DO update set updated_at=EXCLUDED.updated_at, 
	grade_count=EXCLUDED.grade_count, grade_sum=EXCLUDED.grade_sum, lesson_count=EXCLUDED.lesson_count, absent_count=EXCLUDED.absent_count;

-- set 
update user_schools set school_id=null where role_code='admin';


select u.id,u.first_name,us.role_code,s.code,s.id  from users u 
left join user_schools us on us.user_id=u.id left join schools s on s.id=us.school_id 
where us.role_code='organization';

update user_schools us set school_id=(select id from schools where code='etr')
where us.user_id=4540;

select code,name,id from schools where parent_id is null;

delete from schools where code not like '6%' and parent_id is null;
update schools set code='mrs' where code='601';

-- IMPORT
ALTER TABLE schools DROP CONSTRAINT schools_parent_id_fkey;
ALTER TABLE schools DROP CONSTRAINT schools_admin_id_fkey;
COPY schools (id, code, name, full_name, address, email, phone , updated_at, created_at, admin_id, parent_id)
FROM '/csv/school_groups.csv' WITH (FORMAT csv, NULL '\N');

COPY schools (id, code, name, full_name, address, email, phone , updated_at, created_at, admin_id, parent_id)
FROM '/csv/schools.csv' WITH (FORMAT csv, NULL '\N');


COPY users (id, first_name, last_name, middle_name, username, password, status, phone, email, birthday, address, avatar, last_active, updated_at, created_at, gender)
FROM '/csv/users.csv' WITH (FORMAT csv, NULL '\N');

COPY classrooms (id, school_id, name, teacher_id, updated_at, created_at)
FROM '/csv/groups.csv' WITH (FORMAT csv, NULL '\N');

COPY user_schools (user_id, school_id, role_code)
FROM '/csv/user_schools.csv' WITH (FORMAT csv, NULL '\N');


ALTER TABLE user_parents DROP CONSTRAINT user_parents_parent_id_fkey;
ALTER TABLE user_parents DROP CONSTRAINT user_parents_child_id_fkey;
COPY user_parents (parent_id, child_id)
FROM '/csv/user_parents.csv' WITH (FORMAT csv, NULL '\N');

COPY user_classrooms (user_id, classroom_id) -- type, type_key
FROM '/csv/user_classrooms.csv' WITH (FORMAT csv, NULL '\N');
-- ALTER TABLE user_classrooms DROP CONSTRAINT user_classrooms_classroom_id_fkey;
ALTER TABLE user_classrooms DROP CONSTRAINT user_classrooms_user_id_fkey;
COPY user_classrooms (user_id, classroom_id, type, type_key)
FROM '/csv/sub_groups1.csv' WITH (FORMAT csv, NULL '\N');
COPY user_classrooms (user_id, classroom_id, type, type_key)
FROM '/csv/sub_groups2.csv' WITH (FORMAT csv, NULL '\N');

COPY periods (school_id, title, value)
FROM '/csv/periods.csv' WITH (FORMAT csv, NULL '\N', DELIMITER E'\t');

COPY shifts (id, name, school_id, value, created_at, updated_at)
FROM '/csv/schedules.csv' WITH (FORMAT csv, NULL '\N');
update shifts set value=replace(value,'\','"')::json;


ALTER TABLE subjects DROP CONSTRAINT subjects_teacher_id_fkey;
-- ALTER TABLE subjects DROP CONSTRAINT subjects_classroom_id_fkey;
COPY subjects (id, school_id, classroom_id, classroom_type, classroom_type_key, name, full_name, week_hours, teacher_id, created_at, updated_at)
FROM '/csv/subjects.csv' WITH (FORMAT csv, NULL '\N');

COPY timetables (id, school_id, classroom_id, value, shift_id, created_at, updated_at)
FROM '/csv/timetables.csv' WITH (FORMAT csv, NULL '\N');


ALTER TABLE lessons DROP CONSTRAINT lessons_subject_id_fkey;
COPY lessons (id, school_id, subject_id, period_id, period_key, "date", hour_number, 
    type_title, title, content, created_at, updated_at)
FROM '/csv/lessons.csv' WITH (FORMAT csv, NULL '\N');



ALTER TABLE grades DROP CONSTRAINT grades_updated_by_fkey;
ALTER TABLE grades DROP CONSTRAINT grades_created_by_fkey;
ALTER TABLE grades DROP CONSTRAINT grades_student_id_fkey;
ALTER TABLE grades DROP CONSTRAINT grades_lesson_id_fkey;
ALTER TABLE grades ALTER COLUMN updated_by DROP NOT NULL; 
ALTER TABLE grades ALTER COLUMN created_by DROP NOT NULL; 
COPY grades (id, lesson_id, student_id, value, values, comment, reason, 
    created_at, updated_at, deleted_at, created_by, updated_by)
FROM '/csv/grades.csv' WITH (FORMAT csv, NULL '\N');


ALTER TABLE absents DROP CONSTRAINT absents_student_id_fkey; 
ALTER TABLE absents DROP CONSTRAINT absents_lesson_id_fkey; 
ALTER TABLE absents ALTER COLUMN updated_by DROP NOT NULL; 
ALTER TABLE absents ALTER COLUMN created_by DROP NOT NULL; 
COPY absents (id, lesson_id, student_id, comment, reason, 
    created_at, updated_at, deleted_at, created_by, updated_by)
FROM '/csv/absents.csv' WITH (FORMAT csv, NULL '\N');

-- setup: after import reset ids

SELECT setval((select pg_get_serial_sequence('users', 'id')), (SELECT MAX(id) FROM users)+1);
SELECT setval((select pg_get_serial_sequence('classrooms', 'id')), (SELECT MAX(id) FROM classrooms)+1);
SELECT setval((select pg_get_serial_sequence('schools', 'id')), (SELECT MAX(id) FROM schools)+1);
SELECT setval((select pg_get_serial_sequence('subjects', 'id')), (SELECT MAX(id) FROM subjects)+1);
SELECT setval((select pg_get_serial_sequence('shifts', 'id')), (SELECT MAX(id) FROM shifts)+1);
SELECT setval((select pg_get_serial_sequence('timetables', 'id')), (SELECT MAX(id) FROM timetables)+1);
SELECT setval((select pg_get_serial_sequence('periods', 'id')), (SELECT MAX(id) FROM periods)+1);
SELECT setval((select pg_get_serial_sequence('user_classrooms', 'id')), (SELECT MAX(id) FROM user_classrooms)+1);
SELECT setval((select pg_get_serial_sequence('user_schools', 'id')), (SELECT MAX(id) FROM user_schools)+1);
SELECT setval((select pg_get_serial_sequence('user_parents', 'id')), (SELECT MAX(id) FROM user_parents)+1);
SELECT setval((select pg_get_serial_sequence('lessons', 'id')), (SELECT MAX(id) FROM lessons)+1);
SELECT setval((select pg_get_serial_sequence('grades', 'id')), (SELECT MAX(id) FROM grades)+1);
SELECT setval((select pg_get_serial_sequence('absents', 'id')), (SELECT MAX(id) FROM absents)+1);



-- seeds

INSERT INTO schools (code, name, full_name, address, email, phone, test1, created_at, updated_at) VALUES
('ark', 'Arkadag şäheri', 'Arkadag şäheri', NULL, NULL, NULL, 2, '2001-01-01 01:01:01', '2001-01-01 01:01:01'),
('ag', 'Aşgabat şäheri', 'Aşgabat şäheri', NULL, NULL, NULL, 2, '2001-01-01 01:01:01', '2001-01-01 01:01:01'),
('brk', 'Berkararlyk etraby', 'Berkararlyk etraby', NULL, NULL, NULL, 2, '2001-01-01 01:01:01', '2001-01-01 01:01:01'),
('bgt', 'Bagtyýarlyk etraby', 'Bagtyýarlyk etraby', NULL, NULL, NULL, 2, '2001-01-01 01:01:01', '2001-01-01 01:01:01'),
('kpt', 'Köpetdag etraby', 'Köpetdag etraby', NULL, NULL, NULL, 2, '2001-01-01 01:01:01', '2001-01-01 01:01:01'),
('bzm', 'Büzmeýin etraby', 'Büzmeýin etraby', NULL, NULL, NULL, 2, '2001-01-01 01:01:01', '2001-01-01 01:01:01'),
('tjn', 'Tejen şäheri', 'Tejen şäheri', NULL, NULL, NULL, 2, '2001-01-01 01:01:01', '2001-01-01 01:01:01'),
('akb', 'Ak bugdaý etraby', 'Ak bugdaý etraby', NULL, NULL, NULL, 2, '2001-01-01 01:01:01', '2001-01-01 01:01:01'),
('bhr', 'Bäherden etraby', 'Bäherden etraby', NULL, NULL, NULL, 2, '2001-01-01 01:01:01', '2001-01-01 01:01:01'),
('gkd', 'Gökdepe etraby', 'Gökdepe etraby', NULL, NULL, NULL, 2, '2001-01-01 01:01:01', '2001-01-01 01:01:01'),
('kkae', 'Kaka etraby', 'Kaka etraby', NULL, NULL, NULL, 2, '2001-01-01 01:01:01', '2001-01-01 01:01:01'),
( 'kka', 'Kaka şäheri', 'Kaka şäheri', NULL, NULL, NULL, 2, '2001-01-01 01:01:01', '2001-01-01 01:01:01'),
( 'srs', 'Sarahs etraby', 'Sarahs etraby', NULL, NULL, NULL, 2, '2001-01-01 01:01:01', '2001-01-01 01:01:01'),
( 'tjne', 'Tejen etraby', 'Tejen etraby', NULL, NULL, NULL, 2, '2001-01-01 01:01:01', '2001-01-01 01:01:01'),
( 'bbd', 'Babadaýhan etraby', 'Babadaýhan etraby', NULL, NULL, NULL, 2, '2001-01-01 01:01:01', '2001-01-01 01:01:01')
ON CONFLICT (code) DO NOTHING;

update schools
set parent_id=((select id from schools where code='bgt'))
where created_at!='2001-01-01 01:01:01';

INSERT INTO users (first_name, middle_name, last_name, test1, address, phone, birthday, test25, username, email, test2, test3, password, status, test4, test5, test6, created_at, updated_at, test7, avatar, test8, test9, test10, test11, test12, test13, test14, test15, last_active, test16, test17, test18, test19, archived_at, test20, test21, test22, test23, test24) VALUES 
('Agöýli',NULL,'Q','GYL',NULL,'62712204','2020-01-29','male','agoyli','agoyliq@gmail.com',NULL,NULL,'$2y$10$rCNkrG7aPPi8/W8T8wkXZOfTrNxt7SewISYzoj358mgKIcJI5sNzy','active',NULL,NULL,'FqGlb6zWsQMjv7nMFDEilnteC7qZ8PfXvwTKF629HJBkdnZH27jScbCMlS8s','2021-08-05 16:02:11','2023-08-09 06:38:43',38.00,'WQNYfqo1k0lOgeOrfv9P06Nv4wZR0o131UnxTC43.png',NULL,NULL,NULL,0,'142','7','light','tm','2023-08-09 11:38:43','s000101',1,'dTcxeGFwVDBSOjh5aW9IY0N6S0w=','usr8UrNOJRkyWI',NULL,NULL,'super_admin',NULL,'312e9d',26.00),
('Tagandurdy','Seýitnazarowiç','Rejepow','RJPWMYRTSYT',NULL,'61152972','1991-11-21','male','tagandurdy',NULL,NULL,NULL,'$2y$10$rCNkrG7aPPi8/W8T8wkXZOfTrNxt7SewISYzoj358mgKIcJI5sNzy','active',NULL,2,'r4x8v3dBkhPyFRRVBwbhhtsKPVTFhXGiqmbj1ZzI3UPIW0TnCxf9BxUsROqC','2021-05-07 12:22:49','2023-02-11 14:39:05',0.00,'FZZ1UAH8lExCVmITzXyeTHgOqWrBG1kylqLIl1Pp.jpg',NULL,NULL,NULL,0,'2',NULL,NULL,NULL,'2023-02-11 14:39:05','azcm0001',1,'dXdsd0VxdnBPOndVSGpWZ0c4TGc=','usrxX6WS7gx5Kc',NULL,NULL,'teacher',NULL,'1e78e0',0.00),
('Akynyaz','Güýçmyradowic','Alikperow','LKPRWYBLKGY',NULL,'61123016','1978-11-24','female','akynyaz',NULL,NULL,NULL,'$2y$10$rCNkrG7aPPi8/W8T8wkXZOfTrNxt7SewISYzoj358mgKIcJI5sNzy','active',NULL,2,'nnfcZwZxozUSAZIyaWALayhryu6tnf7gWZDMxAaLosOaNMPTuzvYML1zx9bK','2021-05-07 12:22:53','2023-02-21 13:06:58',0.00,'ZmunXROT3pQwo6W6SFFAzcG1UsQF6ehgMfJkF4bH.jpg',NULL,NULL,NULL,0,NULL,NULL,'dark',NULL,'2023-02-21 13:06:58','azct000103',1,'dVN0UjhuMmpEOnBKUklGaXVMSkI=','usr2U2wvouvcqk',NULL,NULL,'teacher',NULL,'47886b',0.00),
('Sohbet','Gylyjowiç','Çörliýew','CRLYWCRLGYL',NULL,'64545533','1991-01-27','male','sohbet',NULL,NULL,NULL,'$2y$10$rCNkrG7aPPi8/W8T8wkXZOfTrNxt7SewISYzoj358mgKIcJI5sNzy','active',NULL,2,NULL,'2021-05-07 12:22:54','2023-01-27 18:24:15',0.00,'WsMHF2GDBnao00s1Zo6jeSyteWCKcABOKwDvQIRb.jpg',NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,'azcm0002',1,NULL,NULL,NULL,NULL,'teacher',NULL,'64b2db',0.00)
ON CONFLICT (username) DO NOTHING;

DELETE from user_schools where 
user_id in
((select id from users where username='agoyli'),
(select id from users where username='tagandurdy'),
(select id from users where username='akynyaz'),
(select id from users where username='sohbet'));

INSERT INTO user_schools (user_id, school_id, role_code) VALUES 
((select id from users where username='agoyli' limit 1), null, 'admin'),
((select id from users where username='akynyaz' limit 1), null, 'admin'),
((select id from users where username='agoyli' limit 1), (select id from schools where code='ag131'), 'principal'),
((select id from users where username='agoyli' limit 1), (select id from schools where code='ag131'), 'teacher'),
((select id from users where username='agoyli' limit 1), (select id from schools where code='ag131'), 'parent'),
((select id from users where username='tagandurdy' limit 1), null, 'admin'),
((select id from users where username='tagandurdy' limit 1), (select id from schools where code='ag131'), 'parent'),
((select id from users where username='tagandurdy' limit 1), (select id from schools where code='ag131'), 'teacher'),
((select id from users where username='akynyaz' limit 1), (select id from schools where code='ag131'), 'parent'),
((select id from users where username='akynyaz' limit 1), (select id from schools where code='ag131'), 'teacher'),
((select id from users where username='sohbet' limit 1), null, 'admin'),
((select id from users where username='sohbet' limit 1), (select id from schools where code='ag131'), 'parent'),
((select id from users where username='sohbet' limit 1), (select id from schools where code='ag131'), 'teacher');

insert into user_parents (parent_id, child_id) values 
((select id from users where username='agoyli'),
(select user_id from user_schools where role_code='student' and school_id=(select id from schools where code='ag131')
  limit 1));



UPDATE subjects SET second_teacher_id=(select id from users where username='agoyli' limit 1) where 
    id=(select id from subjects where name='Matematika' and school_id=(select id from schools where code='ag131') order by id limit 1);
UPDATE subjects SET second_teacher_id=(select id from users where username='tagandurdy' limit 1) where 
    id=(select id from subjects where name='Matematika' and school_id=(select id from schools where code='ag131') order by id  offset 1 limit 1);
UPDATE subjects SET second_teacher_id=(select id from users where username='akynyaz' limit 1) where 
    id=(select id from subjects where name='Matematika' and school_id=(select id from schools where code='ag131') order by id offset 2 limit 1);
UPDATE subjects SET second_teacher_id=(select id from users where username='sohbet' limit 1) where 
    id=(select id from subjects where name='Matematika' and school_id=(select id from schools where code='ag131') order by id offset 3 limit 1);

