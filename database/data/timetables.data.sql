

-- delete from user_schools;
INSERT INTO user_schools (user_uid, school_uid, role_code) VALUES 
((select uid from users where username='agoyli'), (select uid from schools where code='ag21'), 'principal'),
((select uid from users where username='akynyaz'), (select uid from schools where code='ag21'), 'teacher'),
((select uid from users where username='tagan'), (select uid from schools where code='ag21'), 'teacher'),
((select uid from users where username='sohbet'), (select uid from schools where code='ag21'), 'parent'),
((select uid from users where username='okuwcy0'), (select uid from schools where code='ag21'), 'student'),
((select uid from users where username='okuwcy1'), (select uid from schools where code='ag21'), 'student'),
((select uid from users where username='okuwcy2'), (select uid from schools where code='ag21'), 'student'),
((select uid from users where username='okuwcy3'), (select uid from schools where code='ag21'), 'student'),
((select uid from users where username='okuwcy4'), (select uid from schools where code='ag21'), 'student'),
((select uid from users where username='okuwcy5'), (select uid from schools where code='ag21'), 'student');


-- delete from user_classrooms;
INSERT INTO user_classrooms (user_uid, classroom_uid, type, type_key) VALUES 
((select uid from users where username='okuwcy0'), (select uid from classrooms where name='10A'), null, null),
((select uid from users where username='okuwcy1'), (select uid from classrooms where name='10A'), null, null),
((select uid from users where username='okuwcy2'), (select uid from classrooms where name='10A'), null, null),
((select uid from users where username='okuwcy3'), (select uid from classrooms where name='10A'), null, null),
((select uid from users where username='okuwcy4'), (select uid from classrooms where name='10A'), null, null),
((select uid from users where username='agoyli'), (select uid from classrooms where name='10A'), null, null);



-- delete from user_parents;
INSERT INTO user_parents (parent_uid, child_uid) VALUES 
((select uid from users where username='sohbet'), (select uid from users where username='okuwcy0')),
((select uid from users where username='sohbet'), (select uid from users where username='okuwcy4')),
((select uid from users where username='sohbet'), (select uid from users where username='okuwcy5'));



-- delete from shifts;
insert into shifts 
(name, school_uid, value, updated_by_uid) VALUES 
('08:30/12:00 irdenki', (select uid from schools where code='ag21'), 
'[[["08:30", "09:15"],["09:20", "10:05"],["10:10", "10:55"], ["15:10", "16:55"]],[["08:30", "09:15"],["09:20", "10:05"],["10:10", "10:55"], ["11:10", "11:55"]],[["08:30", "09:15"],["09:20", "10:05"],["10:10", "10:55"], ["11:10", "11:55"]],[["08:30", "09:15"],["09:20", "10:05"],["10:10", "10:55"], ["11:10", "11:55"]],[["08:30", "09:15"],["09:20", "10:05"],["10:10", "10:55"], ["11:10", "11:55"]],[["08:30", "09:15"],["09:20", "10:05"],["10:10", "10:55"], ["11:10", "11:55"]]]', null),
('13:30/18:00 agsamky', (select uid from schools where code='ag21'), 
'[[["08:30", "09:15"],["09:20", "10:05"],["10:10", "10:55"], ["11:10", "11:55"]],null,[["08:30", "09:15"],["09:20", "10:05"],["10:10", "10:55"], ["11:10", "11:55"]],null,[["08:30", "09:15"],["09:20", "10:05"],["10:10", "10:55"], ["11:10", "11:55"]],null]', null);



-- delete from subjects;
insert into subjects 
(school_uid, classroom_uid, teacher_uid, week_hours,name) VALUES
((select uid from schools where code='ag21'), (select uid from classrooms where name='10A'), 
(select uid from users where username='akynyaz'), 3, 'Algebra'),
((select uid from schools where code='ag21'), (select uid from classrooms where name='10A'), 
(select uid from users where username='akynyaz'), 3, 'Fizika'),
((select uid from schools where code='ag21'), (select uid from classrooms where name='10A'), 
(select uid from users where username='akynyaz'), 3, 'Informatika'),
((select uid from schools where code='ag21'), (select uid from classrooms where name='10A'), 
(select uid from users where username='tagan'), 3, 'Biologiýa');



-- select * from timetables;
insert into timetables 
(school_uid, classroom_uid, updated_by_uid, shift_uid, value) VALUES 
((select uid from schools where code='ag21'), (select uid from classrooms where name='10A'), 
null, 
(select uid from shifts limit 1), 
json_build_array (
    json_build_array(
        (select uid from subjects where name='Algebra' limit 1), 
        (select uid from subjects where name='Fizika' limit 1), 
        null,
        (select uid from subjects where name='Informatika' limit 1)
    ), 
    NULL,
    json_build_array(
        (select uid from subjects where name='Algebra' limit 1), 
        (select uid from subjects where name='Fizika' limit 1), 
        null,
        (select uid from subjects where name='Informatika' limit 1)
        ), 
    NULL,
    json_build_array(
        (select uid from subjects where name='Algebra' limit 1), 
        (select uid from subjects where name='Fizika' limit 1), 
        null,
        (select uid from subjects where name='Informatika' limit 1)
        ), 
    NULL
    )
)