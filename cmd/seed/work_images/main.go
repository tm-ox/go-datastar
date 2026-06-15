package main

import (
	"log"

	"github.com/tm-ox/go-datastar/internal/db"
)

func main() {
	database, err := db.Open("./data.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	images := map[string][]string{
		"logos": {
			"https://tmox.net/_astro/dark-house.BnYhrUjj_9FW7.webp",
			"https://tmox.net/_astro/gula-2-e.D_-R-pUB_Z1CMTfO.webp",
			"https://tmox.net/_astro/kenji.BZnEly73_GEfr1.webp",
			"https://tmox.net/_astro/odivine-e.DJIr_Zx3_Zrq8Dg.webp",
			"https://tmox.net/_astro/gg.XWEhECPZ_2fkVze.webp",
			"https://tmox.net/_astro/lupa.CNEsFROw_8qP9c.webp",
			"https://tmox.net/_astro/kbloc.BZQN82LF_ZjKQ4x.webp",
			"https://tmox.net/_astro/neon-deli-e.DFNuWtQ2_Z1qGYHr.webp",
			"https://tmox.net/_astro/imogen.CT_g2_mY_Z1nGXtP.webp",
			"https://tmox.net/_astro/mahanyawa-e.D1H6GKXF_13YEPP.webp",
			"https://tmox.net/_astro/aboriginal-owned-e.C3T63JTB_27Qpga.webp",
			"https://tmox.net/_astro/fallon.1j4QSZ--_Z13cA6s.webp",
			"https://tmox.net/_astro/pos-e.BUqUDpQP_28VeSC.webp",
			"https://tmox.net/_astro/boab-e.VLxvwUXC_2tlOl3.webp",
			"https://tmox.net/_astro/hajduk.Da26k67E_1qhAzY.webp",
			"https://tmox.net/_astro/mod.BMwFoM9n_1fncKv.webp",
			"https://tmox.net/_astro/ox-min.BnJAUEyI_ZzruKC.webp",
			"https://tmox.net/_astro/tg.C9FT_U0X_Z38zUn.webp",
		},
		"tg": {
			"https://tmox.net/_astro/tg-gm-login.D2fPD4ob_Z1FzXOr.webp",
			"https://tmox.net/_astro/tg-gm-m-login.BYzGoIZV_1zIl5l.webp",
			"https://tmox.net/_astro/tg-gm-assessments.DkGfMoIL_1JmmFk.webp",
			"https://tmox.net/_astro/tg-gm-m-assessments.BffWc2WW_2cINXH.webp",
			"https://tmox.net/_astro/tg-gm-tg.CxWQOYmD_eS8vf.webp",
			"https://tmox.net/_astro/tg-gm-m-tg.DU8-2nRC_eC6Nr.webp",
			"https://tmox.net/_astro/tg-gm-matrix.DdBz-1SP_1UGfnC.webp",
			"https://tmox.net/_astro/tg-gm-m-matrix.Bgu6yTrD_Z2pF0HH.webp",
		},
		"ntr": {
			"https://tmox.net/_astro/NTR001-Brochure-1.DEt378LL_1K1QhN.webp",
			"https://tmox.net/_astro/NTR001-Brochure-3.DhXItTuE_ZXMIPJ.webp",
			"https://tmox.net/_astro/NTR001-Brochure-6.CBLOsmlK_Z2swCBe.webp",
			"https://tmox.net/_astro/NTR001-Brochure-7.DJttqFbO_2ud0YX.webp",
			"https://tmox.net/_astro/NTR001-Brochure-8.QYG2C-my_Z1mf29s.webp",
		},
		"oxen-free": {
			"https://tmox.net/_astro/ox-logo-min.Yaeft7cE_1XtU4B.webp",
			"https://tmox.net/_astro/ox-logo-max.u6R4bFoV_Z1j7VYG.webp",
			"https://tmox.net/_astro/ox-type.VP1IPEdq_XwROW.webp",
			"https://tmox.net/_astro/ox-member-front.ChSCk5vm_Z15YtKM.webp",
			"https://tmox.net/_astro/ox-member-back.BynmUABX_Z1bezXb.webp",
			"https://tmox.net/_astro/ox-igfeed-ayash2.BsrKP8rM_Z1xVaJb.webp",
			"https://tmox.net/_astro/ox-igstory-ayash2.CGK3zeGS_ZO3C1y.webp",
		},
		"redsky": {
			"https://tmox.net/_astro/RedSky-01.C78wcXrD_Z1eFw1.webp",
			"https://tmox.net/_astro/RedSky-02.CmSU1ZCy_Z11AajD.webp",
			"https://tmox.net/_astro/RedSky-03.1kv5hT3k_2olNjX.webp",
			"https://tmox.net/_astro/RedSky-04.c41qsB_u_Z1Y83de.webp",
			"https://tmox.net/_astro/RedSky-05.CNQjA0kZ_Z1tsfT1.webp",
			"https://tmox.net/_astro/RedSky-06.KVWnu5aX_9cNHv.webp",
			"https://tmox.net/_astro/RedSky-07.Bm-caJvP_Covjp.webp",
		},
		"print": {
			"https://tmox.net/_astro/compass-a4.-vzdV72V_2bRyxX.webp",
			"https://tmox.net/_astro/db-flyer-front.CNURUAyR_2nyyMs.webp",
			"https://tmox.net/_astro/db-flyer-back.BjUbjVtB_Z2dmQE0.webp",
			"https://tmox.net/_astro/dotd-flyer.Bvawrh8z_Z1l8yKs.webp",
			"https://tmox.net/_astro/warrlja-front.DFQk5O1m_Zy99gx.webp",
			"https://tmox.net/_astro/warrlja-back.C0DUo8LC_ZVuU22.webp",
			"https://tmox.net/_astro/MYC-nda-mock.DFdBqtm8_Z1x5cyP.webp",
			"https://tmox.net/_astro/MYC-nda.Crm01r_3_2dH4pY.webp",
			"https://tmox.net/_astro/shorecan-flyer.GlvGKMQQ_Z1SQwVy.webp",
			"https://tmox.net/_astro/SML-outer.-7mVH8Ou_2pwn8J.webp",
			"https://tmox.net/_astro/SML-inner.BSdXMOV1_sfsWH.webp",
			"https://tmox.net/_astro/OX_UnKwn03.Bjuqz8cY_crGzC.webp",
			"https://tmox.net/_astro/OX_FogAcid_2.BmVQm_qe_Z1T31PX.webp",
		},
		"saj": {
			"https://tmox.net/_astro/saj-desktop-1.yzr4oNf7_1lGKVw.webp",
			"https://tmox.net/_astro/saj-desktop-2.C5IC1az2_Z2taXU0.webp",
			"https://tmox.net/_astro/saj-desktop-3.BAzzm79G_Z12E9jp.webp",
			"https://tmox.net/_astro/saj-desktop-4.DtFPS7rF_Z1gphrG.webp",
			"https://tmox.net/_astro/saj-desktop-5.CC84lQ-S_Z21AUKP.webp",
			"https://tmox.net/_astro/saj-elements.DUNCTuAM_Z1AuXIT.webp",
			"https://tmox.net/_astro/saj-mobile-1.y0s5BNc8_Z2hUBDr.webp",
			"https://tmox.net/_astro/saj-mobile-2.CBQLf0yJ_1R3Dg5.webp",
			"https://tmox.net/_astro/saj-mobile-4.D8EQN3jH_17Gafh.webp",
			"https://tmox.net/_astro/saj-mobile-5.B7dTHg5B_Z1jkCur.webp",
		},
		"fredst": {
			"https://tmox.net/_astro/fredst-web-company-2.4FBv51c5_28Q6dR.webp",
			"https://tmox.net/_astro/fredst-web-climate.BRYL60I-_Z1WI0Sd.webp",
			"https://tmox.net/_astro/fredst-web-projects-1.B8QEva8B_Z18GhDM.webp",
			"https://tmox.net/_astro/fredst-web-projects-2.DP1KU6Gd_21VsVN.webp",
			"https://tmox.net/_astro/fredst-web-team.DoYkpJ2O_KC2f9.webp",
			"https://tmox.net/_astro/fredst-web-mobile-2.DhNe71h9_2cBlVH.webp",
			"https://tmox.net/_astro/fredst-web-mobile-3.Bdr-8qwQ_Z2tnmeU.webp",
		},
		"ynk": {
			"https://tmox.net/_astro/ynk-web-dark-home.CPv2KHFJ_1MRK7C.webp",
			"https://tmox.net/_astro/ynk-web-light-home.3G8NYa4R_Z1pyfsw.webp",
			"https://tmox.net/_astro/ynk-web-dark-about.BZdo-lpi_Z2bvncd.webp",
			"https://tmox.net/_astro/ynk-web-dark-images.O5oVe5l-_1K9hw8.webp",
			"https://tmox.net/_astro/ynk-web-light-mobile.CWRBTUK3_Z2sG40n.webp",
			"https://tmox.net/_astro/ynk-web-dark-mobile.ByCqFPul_1zpOQr.webp",
		},
		"mjo": {
			"https://tmox.net/_astro/mjo-web-all.Dbmg_Nop_1Uua11.webp",
			"https://tmox.net/_astro/mjo-web-landing.DLoFeW4D_16YoIb.webp",
			"https://tmox.net/_astro/mjo-web-solutions.DRPmJQEo_1zHaYm.webp",
			"https://tmox.net/_astro/mjo-web-cards.Bbpoc2Hp_6HBKr.webp",
			"https://tmox.net/_astro/mjo-web-contact.CSkcRJwi_27LhDt.webp",
		},
	}

	for slug, urls := range images {
		for i, url := range urls {
			_, err := database.Exec(`
				INSERT OR IGNORE INTO work_images (work_id, url, alt, sort_order)
				SELECT id, ?, '', ? FROM work WHERE slug = ?
			`, url, i, slug)
			if err != nil {
				log.Fatalf("failed to seed image for %s: %v", slug, err)
			}
		}
	}

	log.Println("work_images seed complete")
}
