CREATE TABLE public.admin_to_role_demo (
    admin_id uuid NOT NULL,
    role_id uuid NOT NULL
);
COMMENT ON TABLE public.admin_to_role_demo IS '系统-用户角色关联';
CREATE INDEX admin_to_role_demo_admin_id_role_id_idx ON public.admin_to_role_demo USING btree (admin_id, role_id);
CREATE INDEX admin_to_role_demo_role_id_admin_id_idx ON public.admin_to_role_demo USING btree (role_id, admin_id);
