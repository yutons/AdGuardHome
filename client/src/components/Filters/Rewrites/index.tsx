import React, { Component, Fragment } from 'react';
import { Trans, withTranslation } from 'react-i18next';

import Table from './Table';

import Modal from './Modal';

import Card from '../../ui/Card';

import PageTitle from '../../ui/PageTitle';
import { MODAL_TYPE } from '../../../helpers/constants';
import { RewritesData } from '../../../initialState';
// 定制
import {debounce} from 'lodash';
// 定制

interface RewritesProps {
    t: (...args: unknown[]) => string;
    getRewritesList: () => (dispatch: any) => void;
    toggleRewritesModal: (...args: unknown[]) => unknown;
    addRewrite: (...args: unknown[]) => unknown;
    deleteRewrite: (...args: unknown[]) => unknown;
    updateRewrite: (...args: unknown[]) => unknown;
    rewrites: RewritesData;
}

class Rewrites extends Component<RewritesProps> {
    componentDidMount() {
        this.props.getRewritesList();
    }

    handleDelete = (values: any) => {
        // eslint-disable-next-line no-alert
        if (window.confirm(this.props.t('rewrite_confirm_delete', { key: values.domain }))) {
            this.props.deleteRewrite(values);
        }
    };

    handleSubmit = (values: any) => {
        const { modalType, currentRewrite } = this.props.rewrites;

        if (modalType === MODAL_TYPE.EDIT_REWRITE && currentRewrite) {
            this.props.updateRewrite({
                target: currentRewrite,
                update: values,
            });
        } else {
            this.props.addRewrite(values);
        }
    };

    // 定制
    handleSearchChange = (event) => {
        const params = {param: ''};
        params.param = event.target.value;

        const searchDebounced = debounce(this.props.getRewritesList, 500);
        // @ts-ignore
        searchDebounced(params);
    };
    // 定制

    render() {
        const {
            t,

            rewrites,

            toggleRewritesModal,
        } = this.props;

        const {
            list,
            isModalOpen,
            processing,
            processingAdd,
            processingDelete,
            processingUpdate,
            modalType,
            currentRewrite,
        } = rewrites;

        return (
            // 定制
            <div>
            {/*定制*/}
                {/*定制，增加搜索框*/}
                <div className="page-header page-header--logs">
                    <h1 className="page-title page-title--large">
                        {t('dns_rewrites')}
                    </h1>
                    <div className="d-flex flex-wrap form-control--container">
                        <div className="field__search">
                            <div className="input-group-search input-group-search__icon--magnifier">
                                <svg className="icons icon--24 icon--gray">
                                    <use xlinkHref="#magnifier"></use>
                                </svg>
                            </div>
                            <input type="text" placeholder="请输入主机记录或记录值" onChange={this.handleSearchChange}
                                   className="form-control form-control--search form-control--transparent"
                                   style={{paddingRight: '0px'}}/>
                        </div>
                    </div>
                </div>
                {/*定制*/}
                <Card id="rewrites" bodyType="card-body box-body--settings">
                    <Fragment>
                        <Table
                            list={list}
                            processing={processing}
                            processingAdd={processingAdd}
                            processingDelete={processingDelete}
                            processingUpdate={processingUpdate}
                            handleDelete={this.handleDelete}
                            toggleRewritesModal={toggleRewritesModal}
                        />

                        <button
                            type="button"
                            className="btn btn-success btn-standard mt-3"
                            onClick={() => toggleRewritesModal({ type: MODAL_TYPE.ADD_REWRITE })}
                            disabled={processingAdd}>
                            <Trans>rewrite_add</Trans>
                        </button>

                        <Modal
                            isModalOpen={isModalOpen}
                            modalType={modalType}
                            toggleRewritesModal={toggleRewritesModal}
                            handleSubmit={this.handleSubmit}
                            processingAdd={processingAdd}
                            processingDelete={processingDelete}
                            currentRewrite={currentRewrite}
                        />
                    </Fragment>
                </Card>
            {/*定制*/}
            </div>
            // 定制
        );
    }
}

export default withTranslation()(Rewrites);
