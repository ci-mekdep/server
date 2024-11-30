-- migration: cleaning
update schools set full_name=REPLACE(full_name, '"', '') where full_name like '%"%';
update schools set address=REPLACE(address, '"', '') where address like '%"%';

update users set first_name=REPLACE(first_name, '\"', '') where first_name like '%\"%';
update users set last_name=REPLACE(last_name, '\"', '') where last_name like '%\"%';
update users set middle_name=REPLACE(middle_name, '\"', '') where middle_name like '%\"%';
update users set address=REPLACE(address, '"', '') where address like '%"%';
update users set username=REPLACE(username, '\"', '') where username like '%\"%';
update users set status='active' where status like '[%';
update users set birthday=NULL where birthday like '%0000-00-00%';
update users set phone=REPLACE(REPLACE(REPLACE(REPLACE(phone, '993', ''), '+', ''), ' ', ''), '-', '') 
    where phone like '%-%' or  phone like '% %' or phone like '%+%' or phone like '993%';

update `groups` left join users on groups.teacher_id=users.id set teacher_id=null where users.id is null;

delete `users` from `users` where first_name is null or last_name is null;
delete `users` from `users` left join schools on users.school_id=schools.id where schools.id is null;
delete t1 FROM users t1 JOIN users t2 ON t1.username = t2.username WHERE t1.id > t2.id;
delete `groups` from `groups` left join schools on groups.school_id=schools.id where schools.id is null;
update users set username=concat("u",id) where username is null or username='';

update topics set content=REPLACE(content, '\"', '') where content like '%\"%';
update topics set name=REPLACE(name, '\"', '') where name like '%\"%';



-- EXPORT

SELECT schools.id, slug, name, full_name, address, email, phone, updated_at, created_at, null, school_group_id+(select max(id) from schools)
FROM schools 
left join school_has_admins sa on sa.school_id=schools.id group by schools.id
INTO OUTFILE '/var/lib/mysql-files/schools.csv'
FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';

SELECT id+(select max(id) from schools), slug, name, name, null, null, null, updated_at, created_at, admin_id, null
FROM school_groups 
INTO OUTFILE '/var/lib/mysql-files/school_groups.csv'
FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';

SELECT id, school_id, name_canonical, teacher_id, updated_at, created_at 
FROM `groups` where parent_id is not null 
INTO OUTFILE '/var/lib/mysql-files/groups.csv' 
FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';

SELECT 
id, first_name, last_name, middle_name, username, password, status, phone, email, birthday, address, avatar, last_active, updated_at, created_at,
(CASE 
    WHEN gender = 'female' THEN 2
    WHEN gender = 'male' THEN 1
    WHEN gender = null THEN null
    ELSE 0
END) AS gender
FROM users
INTO OUTFILE '/var/lib/mysql-files/users.csv'
FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';




SELECT id, school_id, (CASE 
    WHEN `role` = 'school_admin' THEN 'principal'
    WHEN `role` = 'principal' THEN 'principal'
    WHEN `role` = 'group_admin' THEN 'organization'
    WHEN `role` = 'analytics_admin' THEN 'organization'
    WHEN `role` = 'super_admin' THEN 'admin'
    ELSE `role`
END) AS role
FROM users where role is not null and role!=''
INTO OUTFILE '/var/lib/mysql-files/user_schools.csv'
FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';


-- SELECT users.id, children.id FROM users 
--     left join parents_has_children pc on pc.parent_id=users.id
--     left join users children on children.id=pc.parent_id
--     where children.id is not null group by users.id, children.id
-- INTO OUTFILE '/var/lib/mysql-files/user_parents.csv'
-- FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';
SELECT parent_id, child_id FROM parents_has_children 
INTO OUTFILE '/var/lib/mysql-files/user_parents.csv'
FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';




SELECT users.id, users.group_id FROM users 
    left join `groups` on groups.id=group_id
    where groups.id is not null and groups.parent_id is not null
group by users.id
INTO OUTFILE '/var/lib/mysql-files/user_classrooms.csv'
FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';


SELECT sg.id, groups.school_id, sg.group_id, sub_groups.name, sg.sub_group_item, subjects.name, subjects.full_name, hours, sg.teacher_id, sg.created_at, sg.updated_at 
FROM subject_groups sg 
    left join `subjects` on subjects.id=subject_id
    left join `groups` on groups.id=group_id
    left join `users` on users.id=sg.teacher_id
    left join `sub_groups` on sub_groups.id=sg.sub_group_id
    where subjects.name is not null and groups.school_id is not null and groups.id is not null
group by sg.id
INTO OUTFILE '/var/lib/mysql-files/subjects.csv'
FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';



SELECT s.id, start_date, school_id,  
JSON_ARRAY( 
JSON_ARRAY(JSON_ARRAY(schedule->'$."1"."1".start_time',schedule->'$."1"."1".end_time'),JSON_ARRAY(schedule->'$."1"."2".start_time',schedule->'$."1"."2".end_time'),JSON_ARRAY(schedule->'$."1"."3".start_time',schedule->'$."1"."3".end_time'),JSON_ARRAY(schedule->'$."1"."4".start_time',schedule->'$."1"."4".end_time'),JSON_ARRAY(schedule->'$."1"."5".start_time',schedule->'$."1"."5".end_time'),JSON_ARRAY(schedule->'$."1"."6".start_time',schedule->'$."1"."6".end_time')), 
JSON_ARRAY(JSON_ARRAY(schedule->'$."2"."1".start_time',schedule->'$."2"."1".end_time'),JSON_ARRAY(schedule->'$."2"."2".start_time',schedule->'$."2"."2".end_time'),JSON_ARRAY(schedule->'$."2"."3".start_time',schedule->'$."2"."3".end_time'),JSON_ARRAY(schedule->'$."2"."4".start_time',schedule->'$."2"."4".end_time'),JSON_ARRAY(schedule->'$."2"."5".start_time',schedule->'$."2"."5".end_time'),JSON_ARRAY(schedule->'$."2"."6".start_time',schedule->'$."2"."6".end_time')), 
JSON_ARRAY(JSON_ARRAY(schedule->'$."3"."1".start_time',schedule->'$."3"."1".end_time'),JSON_ARRAY(schedule->'$."3"."2".start_time',schedule->'$."3"."2".end_time'),JSON_ARRAY(schedule->'$."3"."3".start_time',schedule->'$."3"."3".end_time'),JSON_ARRAY(schedule->'$."3"."4".start_time',schedule->'$."3"."4".end_time'),JSON_ARRAY(schedule->'$."3"."5".start_time',schedule->'$."3"."5".end_time'),JSON_ARRAY(schedule->'$."3"."6".start_time',schedule->'$."3"."6".end_time')), 
JSON_ARRAY(JSON_ARRAY(schedule->'$."4"."1".start_time',schedule->'$."4"."1".end_time'),JSON_ARRAY(schedule->'$."4"."2".start_time',schedule->'$."4"."2".end_time'),JSON_ARRAY(schedule->'$."4"."3".start_time',schedule->'$."4"."3".end_time'),JSON_ARRAY(schedule->'$."4"."4".start_time',schedule->'$."4"."4".end_time'),JSON_ARRAY(schedule->'$."4"."5".start_time',schedule->'$."4"."5".end_time'),JSON_ARRAY(schedule->'$."4"."6".start_time',schedule->'$."4"."6".end_time')), 
JSON_ARRAY(JSON_ARRAY(schedule->'$."5"."1".start_time',schedule->'$."5"."1".end_time'),JSON_ARRAY(schedule->'$."5"."2".start_time',schedule->'$."5"."2".end_time'),JSON_ARRAY(schedule->'$."5"."3".start_time',schedule->'$."5"."3".end_time'),JSON_ARRAY(schedule->'$."5"."4".start_time',schedule->'$."5"."4".end_time'),JSON_ARRAY(schedule->'$."5"."5".start_time',schedule->'$."5"."5".end_time'),JSON_ARRAY(schedule->'$."5"."6".start_time',schedule->'$."5"."6".end_time')), 
JSON_ARRAY(JSON_ARRAY(schedule->'$."6"."1".start_time',schedule->'$."6"."1".end_time'),JSON_ARRAY(schedule->'$."6"."2".start_time',schedule->'$."6"."2".end_time'),JSON_ARRAY(schedule->'$."6"."3".start_time',schedule->'$."6"."3".end_time'),JSON_ARRAY(schedule->'$."6"."4".start_time',schedule->'$."6"."4".end_time'),JSON_ARRAY(schedule->'$."6"."5".start_time',schedule->'$."6"."5".end_time'),JSON_ARRAY(schedule->'$."6"."6".start_time',schedule->'$."6"."6".end_time')) 
) as value , s.created_at, s.updated_at  
FROM schedules s left join schools on schools.id=school_id  
where schools.id is not null 
group by s.id 
INTO OUTFILE '/var/lib/mysql-files/schedules.csv' 
FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';



SELECT t.id, groups.school_id, group_id, 
JSON_ARRAY(
JSON_ARRAY(timetable->'$."1"."1"',timetable->'$."1"."2"',timetable->'$."1"."3"',timetable->'$."1"."4"',timetable->'$."1"."5"',timetable->'$."1"."6"'),
JSON_ARRAY(timetable->'$."2"."1"',timetable->'$."2"."2"',timetable->'$."2"."3"',timetable->'$."2"."4"',timetable->'$."2"."5"',timetable->'$."2"."6"'),
JSON_ARRAY(timetable->'$."3"."1"',timetable->'$."3"."2"',timetable->'$."3"."3"',timetable->'$."3"."4"',timetable->'$."3"."5"',timetable->'$."3"."6"'),
JSON_ARRAY(timetable->'$."4"."1"',timetable->'$."4"."2"',timetable->'$."4"."3"',timetable->'$."4"."4"',timetable->'$."4"."5"',timetable->'$."4"."6"'),
JSON_ARRAY(timetable->'$."5"."1"',timetable->'$."5"."2"',timetable->'$."5"."3"',timetable->'$."5"."4"',timetable->'$."5"."5"',timetable->'$."5"."6"'),
JSON_ARRAY(timetable->'$."6"."1"',timetable->'$."6"."2"',timetable->'$."6"."3"',timetable->'$."6"."4"',timetable->'$."6"."5"',timetable->'$."6"."6"')
) as value
, schedule_id, t.created_at, t.updated_at 
FROM timetables t  
    left join `groups` on groups.id=group_id
    left join `schedules` on schedules.id=schedule_id
    left join schools on schools.id=schedules.school_id 
    where schools.id is not null and schedules.id is not null and groups.id is not null and groups.parent_id is not null
group by t.id
INTO OUTFILE '/var/lib/mysql-files/timetables.csv'
FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';



SELECT id, 
    '"2023/24"','"[[""2023-09-01"",""2023-10-22""],[""2023-10-31"",""2023-12-29""],[""2024-01-12"",""2024-03-19""],[""2024-03-29"",""2024-05-25""]]"'
FROM schools
INTO OUTFILE '/var/lib/mysql-files/periods.csv'
LINES TERMINATED BY '\n';


select CAST(student.value AS SIGNED) AS integer_value , sg.group_id,  sg.name, 1
from sub_groups sg
right join `groups` on groups.id=sg.group_id,
JSON_TABLE(
    sg.students,
    '$[0][*]' COLUMNS (
        value JSON PATH '$'
    )
) AS student
INTO OUTFILE '/var/lib/mysql-files/sub_groups1.csv'
FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';

select CAST(student.value AS SIGNED) AS integer_value , group_id,  sg.name, 2
from sub_groups sg
right join `groups` on groups.id=sg.group_id,
JSON_TABLE(
    sg.students,
    '$[1][*]' COLUMNS (
        value JSON PATH '$'
    )
) AS student
INTO OUTFILE '/var/lib/mysql-files/sub_groups2.csv'
FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';



SELECT lessons.id, `groups`.school_id, subject_group_id, null, null, `date`, `index`,  
lesson_type, max(topics.name), max(topics.content), lessons.created_at, lessons.updated_at 
FROM lessons 
left join topics on topics.lesson_id=lessons.id 
left join subject_groups on subject_groups.id=lessons.subject_group_id  
left join `groups` on `groups`.id=subject_groups.group_id 
where deleted_at is null and `groups`.school_id is not null
group by lessons.id
INTO OUTFILE '/var/lib/mysql-files/lessons.csv'
FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';



SELECT id, lesson_id, student_id, value, null, comment, null, 
created_at, updated_at, deleted_at, created_by, updated_by
FROM marks
INTO OUTFILE '/var/lib/mysql-files/grades.csv'
FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';



SELECT id, lesson_id, student_id, comment, reason, 
created_at, updated_at, deleted_at, null, null
FROM mark_absents
INTO OUTFILE '/var/lib/mysql-files/absents.csv'
FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n';




delete duplicate username set null
delete non exist classroom.teacher_id set null


